# Vinom  

Vinom is a secure and scalable system designed for managing real-time communication and game matchmaking. It implements a secure protocol using DTLS and provides robust architecture for authentication, user management, and game server communication.  

## Features  

- **UDP Socket Manager**: Implements secure communication using DTLS for low-latency interactions.  
- **Redis-based Matchmaking**: Efficiently pairs users for games based on predefined criteria.  
- **Game Server Management**: Handles authenticated user requests and manages gameplay sessions with countdowns.  
- **Game Client Management**: terminal.  

## Protocol Overview  

### Authentication Workflow  
1. **Client Fetches Server Public Key**: The client retrieves the server's public key.  
2. **Client Hello**: The client sends a `hello` message containing a random value and an AES-CBC key, encrypted with the server's public key.  
3. **Server Hello Verify**: The server responds with a `helloverify` message containing a cookie HMAC, using the AES-CBC key from the client.  
4. **Client Verification**: The client sends a `hello` message again with the cookie HMAC, AES-CBC key, and a verification token.  
5. **Server Hello**: The server completes the handshake by sending a session ID.  

### Game Server Workflow  
1. Receives authenticated users from the authentication server.  
2. Verifies that users are matched for the current game.  
3. Handles requests from authenticated users sent via the UDP server.  
4. Manages countdowns and game session state.  

## Technologies  

- **Proto3**: For defining structured data communication.  
- **Redis**: For matchmaking and session management.  
- **PostgreSQL**: For persistent data storage.  
- **Docker & Kubernetes**: For containerization and deployment.  
---

![goofer](./assets/logo/gopher-dance-long-3x.gif)

