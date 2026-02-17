#  Real-time Cinema Booking System

Client (Next.js)
    ↓ HTTP + WebSocket
Go Backend (Gin)
    ├── MongoDB (Source of Truth)
    ├── Redis (Distributed Lock)
    ├── Background Worker (Auto Revert)
    └── WebSocket Hub (Real-time Update)

## Tech Stack
Backend:
- Go (Gin)
- MongoDB
- Redis
- Gorilla WebSocket

Frontend:
- Next.js
- TailwindCSS

Infrastructure:
- Docker & Docker Compose

## Distributed Lock Strategy

- Redis SET NX EX used for seat locking (5-minute TTL)
- MongoDB stores seat status (AVAILABLE, LOCKED, BOOKED)
- Lock owner validation before confirmation
- Auto-revert worker restores seat when lock expires

# Seat State 
AVAILABLE
   ↓ Lock
LOCKED
   ↓ Confirm
BOOKED

TTL Expired:
LOCKED → AVAILABLE 

## Concurrency Handling

- Atomic confirmation using conditional Mongo update
- Lock owner validation before booking
- Background worker ensures consistency with Redis TTL
- Prevents double booking in concurrent scenarios

## Real-time Updates

WebSocket is used to broadcast seat status changes:
- seat_locked
- seat_booked
- seat_released

Clients automatically update UI without refreshing.

## Run Locally

```bash
docker compose up --build