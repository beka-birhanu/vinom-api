package udp

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/beka-birhanu/vinom-api/config"
	"github.com/beka-birhanu/vinom-api/service/i"
	"github.com/google/uuid"
)

// ClientRequestHandler is called when an authenticated client request is received (i.e., the client has passed the DTLS handshake).
type ClientRequestHandler func(uuid.UUID, byte, []byte)

// ClientRegisterHandler is called when a client is registerd into a session after being authenticated.
type ClientRegisterHandler func(uuid.UUID)

type ServerOption func(*ServerSocketManager)

// Custom error types
var (
	ErrInvalidRecordType            = errors.New("invalid record type")
	ErrInsecureEncryptionKeySize    = errors.New("insecure encryption key size")
	ErrClientSessionNotFound        = errors.New("client session not found")
	ErrClientAddressIsNotRegistered = errors.New("client address is not registered")
	ErrClientNotFound               = errors.New("client not found")
	ErrMinimumPayloadSizeLimit      = errors.New("minimum payload size limit")
	ErrMaximumPayloadSizeLimit      = errors.New("maximum payload size limit")
	ErrClientCookieIsInvalid        = errors.New("client cookie is invalid")
	ErrInvalidPayloadBodySize       = errors.New("invalid payload body size")
)

const (
	ClientHelloRecordType byte = 1 << iota
	HelloVerifyRecordType
	ServerHelloRecordType
	PingRecordType
	PongRecordType
	UnAuthenticated

	defaultReadBufferSize int = 2048

	minimumPayloadSize  int = 3
	insecureSymmKeySize int = 32 // A symmetric key smaller than 256 bits is insecure. 256 bits = 32 bytes in size.
)

// Incoming bytes are parsed into the record struct
type record struct {
	Type byte
	Body []byte
}

// rawRecord is sent to the rawRecords channel when a new payload is received
type rawRecord struct {
	payload []byte
	addr    *net.UDPAddr
}

// client represents an authenticated UDP client
type client struct {
	ID            uuid.UUID    // ID provided by the authenticator.
	sessionID     []byte       // Session ID is a secret byte array that indicates the client has completed the handshake process. The client must prepend these bytes to the start of each record body before encryption.
	addr          *net.UDPAddr // UDP address of the client.
	eKey          []byte       // Client encryption key for encrypting and decrypting record bodies with the symmetric encryption algorithm.
	lastHeartbeat time.Time    // Last time a record was received from the client.

	sync.Mutex
}

// ServerSocketManager is a UDP socket manager that accepts connections, performs the DTLS handshake, and processes client requests after validation.
type ServerSocketManager struct {
	readBufferSize          int                   // Maximum buffer size for incoming bytes.
	heartbeatExpiration     time.Duration         // Expiration time of the last heartbeat before requiring reauthentication.
	conn                    *net.UDPConn          // Connection to listen to.
	authenticator           i.PlayerAuthenticator // An implementation of Authenticator to authenticate client tokens and return user identifiers.
	encoder                 i.SocketEncoder       // An implementation of Encoder to encode and decode messages.
	HMAC                    i.HMAC
	asymmCrypto             i.Asymmetric          // An implementation of asymmetric encryption.
	symmCrypto              i.Symmetric           // An implementation of symmetric encryption.
	onCustomClientRequest   ClientRequestHandler  // Request handler function called when an authenticated client sends a request.
	onClientRegister        ClientRegisterHandler // Request handler function called when a client completes the DTLS handshake.
	clients                 map[uuid.UUID]*client // Map of clients indexed by their identifier.
	clientsLock             sync.RWMutex          // Read-write lock for accessing the clients map.
	garbageCollectionTicker *time.Ticker          // Client garbage collection ticker.
	garbageCollectionStop   chan bool             // Channel to signal stopping the client garbage collector.
	sessionManager          *SessionManager       // The session manager generates cookies and session IDs.
	rawRecords              chan rawRecord        // Channel for raw records.
	logger                  *log.Logger           // Logger.
	stop                    chan bool             // Channel to signal stopping the server.
	wg                      *sync.WaitGroup       // WaitGroup to manage server goroutines.
}

// ServerConfig is a struct used to pass the required parameters to initialize a new SocketManager
type ServerConfig struct {
	ListenAddr    *net.UDPAddr          // UDP address to listen on.
	Authenticator i.PlayerAuthenticator // An implementation of Authenticator to authenticate client tokens and return user identifiers.
	AsymmCrypto   i.Asymmetric          // An implementation of asymmetric encryption.
	SymmCrypto    i.Symmetric           // An implementation of symmetric encryption.
	Encoder       i.SocketEncoder       // An implementation of Encoder to encode and decode messages.
	HMAC          i.HMAC
}

// NewServerSocketManager initializes a new SocketManager instance with the given configuration and options
func NewServerSocketManager(c ServerConfig, options ...ServerOption) (*ServerSocketManager, error) {
	conn, err := net.ListenUDP("udp", c.ListenAddr)
	if err != nil {
		return nil, err
	}

	s := &ServerSocketManager{
		conn: conn,

		clients:     make(map[uuid.UUID]*client),
		clientsLock: sync.RWMutex{},

		garbageCollectionStop: make(chan bool, 1),
		stop:                  make(chan bool, 1),

		rawRecords: make(chan rawRecord),

		wg: &sync.WaitGroup{},
	}

	// Run optional configurations
	for _, opt := range options {
		opt(s)
	}

	if s.readBufferSize == 0 {
		s.readBufferSize = defaultReadBufferSize
	}

	s.sessionManager, err = NewSessionManager(c.HMAC)
	if err != nil {
		return nil, err
	}

	if s.logger == nil {
		// Discard logging if no logger is set
		s.logger = log.New(io.Discard, "", 0)
	}

	// Must provide required parameters
	s.asymmCrypto = c.AsymmCrypto
	s.symmCrypto = c.SymmCrypto
	s.authenticator = c.Authenticator
	s.encoder = c.Encoder
	s.HMAC = c.HMAC

	return s, nil
}

// Serve starts listening to the UDP port for incoming bytes & then sends payload and sender address into the rawRecords channel if no error is found
func (s *ServerSocketManager) Serve() {
	// If heartbeatExpiration is provided spawn garbage collection routine
	if s.heartbeatExpiration > 0 {
		if s.garbageCollectionTicker != nil {
			s.garbageCollectionStop <- true
			s.garbageCollectionTicker.Stop()
		}
		s.garbageCollectionTicker = time.NewTicker(s.heartbeatExpiration)
		s.garbageCollectionStop = make(chan bool, 1)
		go s.clientGarbageCollection()
	}

	s.rawRecords = make(chan rawRecord)
	go s.handleRawRecords()

	err := s.conn.SetReadDeadline(time.Time{})
	if err != nil {
		s.logger.Printf("%s[ERROR]%s resetting connection deadline: %s", config.LogErrorColor, config.LogColorReset, err)
	}
	s.stop = make(chan bool, 1) // reset the stop channel
	s.logger.Printf("%s[INFO]%s server listening on udp address: %v", config.LogInfoColor, config.LogColorReset, s.conn.LocalAddr().String())
	for {
		select {
		case <-s.stop:
			return
		default:
			buf := make([]byte, s.readBufferSize+1) // Intentionally create more space than allowed for checking
			n, addr, err := s.conn.ReadFromUDP(buf)
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					continue
				}

				s.logger.Printf("%s[ERROR]%s while reading from udp: %s", config.LogErrorColor, config.LogColorReset, err)
				continue
			} else if n > s.readBufferSize {
				s.logger.Printf("%s[ERROR]%s while reading from udp: %s", config.LogErrorColor, config.LogColorReset, ErrMaximumPayloadSizeLimit)
				continue
			}
			s.rawRecords <- rawRecord{
				payload: buf[0:n],
				addr:    addr,
			}
		}
	}
}
func (s *ServerSocketManager) Stop() {
	s.logger.Printf("%s[INFO]%s server stoping gracefuly...", config.LogErrorColor, config.LogColorReset)
	defer s.logger.Printf("%s[INFO]%s server stoped", config.LogErrorColor, config.LogColorReset)

	_ = s.conn.SetReadDeadline(time.Unix(0, 1))
	s.stop <- true
	s.garbageCollectionStop <- true
	s.garbageCollectionTicker.Stop()
	close(s.rawRecords)
	s.wg.Wait()
}

// clientGarbageCollection continuously monitors the connected clients
// and removes any client whose last heartbeat exceeds the heartbeat expiration duration.
func (s *ServerSocketManager) clientGarbageCollection() {
	for {
		select {
		case <-s.garbageCollectionStop: // Assuming the routine writing to stop stops the ticker.
			break
		case <-s.garbageCollectionTicker.C:
			for _, c := range s.clients {
				if time.Now().After(c.lastHeartbeat.Add(s.heartbeatExpiration)) {
					s.clientsLock.Lock()
					delete(s.clients, c.ID)
					s.clientsLock.Unlock()
				}
			}
		}
	}
}

func (s *ServerSocketManager) handleRawRecords() {
	for r := range s.rawRecords {
		s.handleRawRecord(r.payload, r.addr)
	}
}

func (s *ServerSocketManager) handleRawRecord(payload []byte, addr *net.UDPAddr) {
	if len(payload) < minimumPayloadSize {
		s.logger.Printf("%s[ERROR]%s while receving raw record: %s", config.LogErrorColor, config.LogColorReset, ErrMinimumPayloadSizeLimit)
	}

	record, err := parseRecord(payload)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while parsing record: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	switch record.Type {
	case ClientHelloRecordType:
		s.handleHandshakeRecord(record, addr)
	case PingRecordType:
		s.handlePingRecord(record, addr)
	default:
		s.handleCustomRecord(record, addr)
	}
}

// handleHandshakeRecord handles the handshake process for a client connection.
//
// The handshake process includes the following steps:
//  1. The client sends a HandshakeClientHello record encrypted with the server's public key.
//     This record contains the client's encryption key.
//  2. If the ClientHello is valid, the server generates a unique cookie for the client's address,
//     encrypts it with the client key, and sends it back as a HelloVerify record.
//  3. The client responds with a HandshakeClientHelloVerify request containing the generated cookie
//     and token to prove the sender's address is valid.
//  4. The server validates the HelloVerify record, authenticates the client's token, and if valid,
//     generates a session ID. The session ID is encrypted and sent back as a ServerHello record.
//
// Post-registration, clients must prepend the Session ID to the record body (unencrypted bytes),
// then encrypt them and compose the record.
func (s *ServerSocketManager) handleHandshakeRecord(r *record, addr *net.UDPAddr) {
	payload, err := s.asymmCrypto.Decrypt(r.Body)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while decrypting record body: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	handshake, err := s.encoder.UnmarshalHandshake(payload)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while unmarshaling client hello record: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}
	// First client hello
	if len(handshake.GetCookie()) == 0 {
		s.sayHelloVerify(addr, handshake)
	} else { // Second client hello
		s.sayServerHello(addr, handshake)
	}
}

// handlePingRecord handles ping record and sends pong response
func (s *ServerSocketManager) handlePingRecord(r *record, addr *net.UDPAddr) {
	cl, err := s.findClientWithAddr(addr)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while authenticating ping record: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	pong := s.encoder.NewPongRecord()
	pong.SetReceivedAt(time.Now().UnixNano() / int64(time.Millisecond))

	pingPayload, err := s.symmCrypto.Decrypt(r.Body, cl.eKey)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while decrypting ping record: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	sessionID, body, err := splitSessionIDAndBody(pingPayload, len(cl.sessionID))
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while parsing session id for ping: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	if !bytes.Equal(sessionID, cl.sessionID) {
		s.logger.Printf("%s[ERROR]%s while validating session id for ping: %s", config.LogErrorColor, config.LogColorReset, ErrClientSessionNotFound)
		s.unAuthenticated(addr)
		return
	}

	pingRecord, err := s.encoder.UnmarshalPing(body)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while unmarshaling ping record: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	pong.SetPingSentAt(pingRecord.GetSentAt())
	pong.SetSentAt(time.Now().UnixNano() / int64(time.Millisecond))

	pongPayload, err := s.encoder.MarshalPong(pong)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while marshaling pong record: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	err = s.sendToClient(cl, PongRecordType, pongPayload)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while sending pong recored: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	cl.Lock()
	cl.lastHeartbeat = time.Now()
	cl.Unlock()
}

// handleCustomRecord handle custom record with authorizing the record and call the handler func if is set
func (s *ServerSocketManager) handleCustomRecord(r *record, addr *net.UDPAddr) {
	cl, err := s.findClientWithAddr(addr)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while authenticating custom record: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	payload, err := s.symmCrypto.Decrypt(r.Body, cl.eKey)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while decrypting custom record: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	sessionID, body, err := splitSessionIDAndBody(payload, len(cl.sessionID))
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while parsing session id for custom: %s", config.LogErrorColor, config.LogColorReset, err)
		s.unAuthenticated(addr)
		return
	}

	if !bytes.Equal(sessionID, cl.sessionID) {
		s.logger.Printf("%s[ERROR]%s while validating session id for ping: %s", config.LogErrorColor, config.LogColorReset, ErrClientSessionNotFound)
		return
	}

	go s.onCustomClientRequest(cl.ID, r.Type, body)
}

// sayHelloVerify generates and sends a HelloVerify record to the client.
func (s *ServerSocketManager) sayHelloVerify(addr *net.UDPAddr, h i.HandshakeRecord) {
	cookie := s.sessionManager.GetAddrCookieHMAC(addr, h.GetRandom())
	if len(h.GetKey()) < insecureSymmKeySize {
		s.logger.Printf("%s[ERROR]%s while handling hello verify record: %s", config.LogErrorColor, config.LogColorReset, ErrInsecureEncryptionKeySize)
		return
	}

	helloVerify := s.encoder.NewHandshakeRecord()
	helloVerify.SetCookie(cookie)
	helloVerify.SetTimestamp(time.Now().UnixNano() / int64(time.Millisecond))

	helloVerifyPayload, err := s.encoder.MarshalHandshake(helloVerify)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while marshaling hello verify record: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	helloVerifyPayload, err = s.symmCrypto.Encrypt(helloVerifyPayload, h.GetKey())
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while encrypting HelloVerify: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}
	helloVerifyMessage := append([]byte{HelloVerifyRecordType}, helloVerifyPayload...)
	err = s.sendToAddr(addr, helloVerifyMessage)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while sending HelloVerify record to the client: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}
}

// sayServerHello processes the second client handshake and completes the handshake process.
func (s *ServerSocketManager) sayServerHello(addr *net.UDPAddr, h i.HandshakeRecord) {
	cookie := s.sessionManager.GetAddrCookieHMAC(addr, h.GetRandom())
	if !s.HMAC.Compare(h.GetCookie(), cookie) {
		s.logger.Printf("%s[ERROR]%s while validating HelloVerify record cookie: %s", config.LogErrorColor, config.LogColorReset, ErrClientCookieIsInvalid)
		return
	}
	if len(h.GetKey()) < insecureSymmKeySize {
		s.logger.Printf("%s[ERROR]%s while handling HelloVerify recored: %s", config.LogErrorColor, config.LogColorReset, ErrInsecureEncryptionKeySize)
		return
	}

	var token []byte
	var err error
	if len(h.GetToken()) > 0 {
		token, err = s.symmCrypto.Decrypt(h.GetToken(), h.GetKey())
		if err != nil {
			s.logger.Printf("%s[ERROR]%s while decrypting HelloVerify record token: %s", config.LogErrorColor, config.LogColorReset, err)
			return
		}
	}

	ID, err := s.authenticator.Authenticate(token)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while authenticating client token: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	client, err := s.registerClient(addr, ID, h.GetKey())
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while registering client: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	serverHello := s.encoder.NewHandshakeRecord()
	serverHello.SetSessionId(client.sessionID)
	serverHello.SetTimestamp(time.Now().UnixNano() / int64(time.Millisecond))

	serverHelloPayload, err := s.encoder.MarshalHandshake(serverHello)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while marshaling server hello record: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	err = s.sendToClient(client, ServerHelloRecordType, serverHelloPayload)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while sending server hello: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}

	s.logger.Printf("%s[INFO]%s accepted connection with client: %s", config.LogInfoColor, config.LogColorReset, ID)
}

// registerClient generates a new session ID & registers an address with client ID & encryption key as a Client
func (s *ServerSocketManager) registerClient(addr *net.UDPAddr, ID uuid.UUID, eKey []byte) (*client, error) {
	sessionID, err := s.sessionManager.GenerateSessionID(addr, ID)
	if err != nil {
		return nil, err
	}

	cl := &client{
		ID:            ID,
		sessionID:     sessionID,
		addr:          addr,
		eKey:          eKey,
		lastHeartbeat: time.Now(),
	}

	s.clientsLock.Lock()
	s.clients[ID] = cl
	s.clientsLock.Unlock()

	if s.onClientRegister != nil {
		s.onClientRegister(cl.ID)
	}

	return cl, nil
}

// findClientWithAddr finds a registerd client with given addr.
// read locks client lock.
func (s *ServerSocketManager) findClientWithAddr(a *net.UDPAddr) (*client, error) {
	var client *client
	var err error
	s.clientsLock.RLocker().Lock()
	defer s.clientsLock.RLocker().Unlock()

	for _, cl := range s.clients {
		if net.IP.Equal(cl.addr.IP, a.IP) && cl.addr.Port == a.Port {
			client = cl
			break
		}
	}

	if client == nil {
		err = ErrClientAddressIsNotRegistered
	}

	return client, err
}

// BroadcastToClients broadcasts bytes to all registered Clients
func (s *ServerSocketManager) BroadcastToClients(clientIDs []uuid.UUID, typ byte, payload []byte) {
	for _, clID := range clientIDs {
		cl, ok := s.clients[clID]
		if !ok {
			s.logger.Printf("%s[ERROR]%s while writing to the clients: client not found", config.LogErrorColor, config.LogColorReset)
			continue
		}

		s.wg.Add(1)
		go func(c *client) {
			defer s.wg.Done()
			err := s.sendToClient(c, typ, payload)
			if err != nil {
				s.logger.Printf("%s[ERROR]%s while writing to the client: %s", config.LogErrorColor, config.LogColorReset, err)
			}
		}(cl)
	}
}

// sends a record byte array to the Client. the record type is prepended to the record body as a byte
func (s *ServerSocketManager) SendToClient(clientID uuid.UUID, typ byte, payload []byte) error {
	s.clientsLock.RLock()
	client, found := s.clients[clientID]
	if !found {
		return ErrClientNotFound
	}
	s.clientsLock.RUnlock()

	return s.sendToClient(client, typ, payload)
}

// sends a record byte array to the Client. the record type is prepended to the record body as a byte
func (s *ServerSocketManager) sendToClient(client *client, typ byte, payload []byte) error {
	payload, err := s.symmCrypto.Encrypt(payload, client.eKey)
	if err != nil {
		return err
	}
	payload = append([]byte{typ}, payload...)
	return s.sendToAddr(client.addr, payload)
}

// sends a message byte array to the address given.
func (s *ServerSocketManager) sendToAddr(addr *net.UDPAddr, message []byte) error {
	_, err := s.conn.WriteToUDP(message, addr)
	return err
}

// unAuthenticated sends unAuthenticated recorde to client.
// Indicating handshake required.
func (s *ServerSocketManager) unAuthenticated(addr *net.UDPAddr) {
	payload := []byte{UnAuthenticated}
	err := s.sendToAddr(addr, payload)
	if err != nil {
		s.logger.Printf("%s[ERROR]%s while sending UnAuthenticated record to the client: %s", config.LogErrorColor, config.LogColorReset, err)
		return
	}
}

// parseRecord parses a byte slice into a record struct.
//
// The input format depends on the record type:
//   - For most records: [type, body]
//   - For specific types (e.g., HandshakeClientHelloRecordType): [type, bodysize (2 bytes), body, extra]
func parseRecord(r []byte) (*record, error) {
	if len(r) < 2 {
		return nil, ErrInvalidPayloadBodySize
	}

	return &record{
		Type: r[0],
		Body: r[1:],
	}, nil
}

// SplitSessionIDAndBody splits sessionID and body from payload
func splitSessionIDAndBody(payload []byte, sIDLength int) ([]byte, []byte, error) {
	if len(payload) < sIDLength {
		return nil, nil, ErrInvalidPayloadBodySize
	}

	return payload[:sIDLength], payload[sIDLength:], nil
}

// ServerWithClientRequestHandler sets a callback function to handle custom record types received from the client
func ServerWithClientRequestHandler(f ClientRequestHandler) ServerOption {
	return func(s *ServerSocketManager) {
		s.onCustomClientRequest = f
	}
}

// ServerWithClientRegisterHandler sets a callback function to handle client registration after the DTLS handshake
func ServerWithClientRegisterHandler(f ClientRegisterHandler) ServerOption {
	return func(s *ServerSocketManager) {
		s.onClientRegister = f
	}
}

// ServerWithHeartbeatExpiration sets the server heartbeat expiration option
func ServerWithHeartbeatExpiration(t time.Duration) ServerOption {
	return func(s *ServerSocketManager) {
		s.heartbeatExpiration = t
	}
}

// ServerWithReadBufferSize sets the read buffer size option
func ServerWithReadBufferSize(i int) ServerOption {
	return func(s *ServerSocketManager) {
		s.readBufferSize = i
	}
}

// ServerWithLogger sets the logger
func ServerWithLogger(l *log.Logger) ServerOption {
	return func(s *ServerSocketManager) {
		s.logger = l
	}
}

func (s *ServerSocketManager) SetClientRequestHandler(f func(uuid.UUID, byte, []byte)) {
	s.onCustomClientRequest = f
}

func (s *ServerSocketManager) SetClientRegisterHandler(f func(uuid.UUID)) {
	s.onClientRegister = f
}

func (s *ServerSocketManager) SetClientAuthenticator(a i.PlayerAuthenticator) {
	s.authenticator = a
}

func (s *ServerSocketManager) GetPublicKey() []byte {
	return s.asymmCrypto.GetPublicKey()
}

func (s *ServerSocketManager) GetAddr() string {
	return s.conn.LocalAddr().String()
}
