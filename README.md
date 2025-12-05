# RPC Chat System with Real-time Broadcasting

A Go-based distributed chat system using RPC with real-time broadcasting capabilities.

## Features

- Real-time message broadcasting using goroutines and channels
- User join/leave notifications
- No self-echo for messages
- Thread-safe client management using sync.RWMutex
- Full chat history for new clients
- Concurrent client handling

## Installation & Running

### 1. Clone and setup
```bash
git clone <https://github.com/EngZakariaReda/go-rpc-chat.git>
cd go-rpc-chat