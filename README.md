#  Real-time Cinema Booking System

# System Architecture Diagram

Frontend (Next.js)
    ↓ Google Login (Firebase)
    ↓ ได้ Firebase ID Token
    ↓ ส่ง Authorization: Bearer <firebase_token>
Backend
    ↓ Verify Firebase Token
    ↓ Extract user_id
    ↓ ใช้ user_id กับ booking
    ↓ MongoDB (Source of Truth)
    ↓ Redis (Distributed Lock)
    ↓ Background Worker (Auto Revert)
    ↓ WebSocket Hub (Real-time Update)

------------------------------------------------------

# Tech Stack
Backend:
- Go (Gin)
- MongoDB
- Redis
- Gorilla WebSocket

Frontend:
- Next.js
- TailwindCSS
- TypeScript
- WebSocket (real-time)

Data:
- MongoDB

Concurrency:
- Redis

Infrastructure:
- Docker & Docker Compose

-------------------------------------------------------

# Booking Flow

step 1 (User Login)
-   User logs in via Google (Firebase Auth)
-   Frontend receives ID Token
-   Token sent to backend via Authorization header

step 2 (Fetch Shows)
- Frontend calls:
- GET /shows
- Backend returns list of available movie shows.

step 3 (Fetch Seats)
- GET /seats?show_id=show1
- Backend filters seats by show_id.

step 4 (Lock Seat)
- POST /seats/lock
- Backend:  Checks seat availability -> Acquires Redis lock -> Updates
seat  LOCKED -> Broadcasts via WebSocket -> Inserts audit log

step 5 (Payment)
- POST /payment/success
- Backend:  Validates Redis lock owner -> Updates seat LOCKED(BOOKED) ->Inserts booking record  -> Releases Redis lock -> Broadcasts update ->Logs event

step 6 (Admin Monitoring)
- Admin can:  View bookings , Filter by show_id , View audit logs

-------------------------------------------------------------------

# Redis Lock Strategy

- Redis SET NX EX used for seat locking (5-minute TTL)
- MongoDB stores seat status (AVAILABLE, LOCKED, BOOKED)
- Lock owner validation before confirmation
- Auto-revert worker restores seat when lock expires

------------------------------------------------------------------

#  Message Queue Usage
- WebSocket is used as a real-time event distributor.
- Events: - seat_locked - seat_booked
-  Flow: User locks seat → Backend updates DB → WebSocket broadcast → All
clients refresh UI

-------------------------------------------------------------------------

#  How to Run the System
-  Start MongoDB & Redis & Backend
                |
  docker compose up --build
  Backend runs on: http://localhost:8080

-  Start Frontend
        |
    cd frontend
    npm install
    npm run dev
  Frontend runs on: http://localhost:3000

-----------------------------------------------------------

#  Assumptions & Trade-offs
##  Assumptions
- User Authentication: Assumes all users possess a Google account
- Network Stability: Assumes a reasonably stable internet connection to ensure seamless real-time seat updates via WebSockets.
- Server-side Time Reference: Assumes that the reservation expiration and countdown timers are synchronized based on the server's system clock.

## Trade-offs
- User Authentication (Google Account Only)
 |_Pros: Rapid Development: Drastically reduces development time by 
 |    offloading auth logic to Firebase.
 |_Cons: Accessibility Barrier: Excludes potential users who do not have or 
       do not wish to use a Google account.

- Network Stability (WebSockets for Real-time Updates)
|_Pros: Seamless UX: Users see seat status changes instantly without 
|      manually refreshing the page.
|_Cons: High Resource Overhead: Maintaining persistent open connections 
      increases memory usage on both the client and server.

- Server-side Time Reference (Countdown Sync)
|_Pros: Anti-Cheating: Prevents users from bypassing seat locks by manually 
|      altering their local system clock.
|_Cons: Perceived Latency: Users with high network ping may notice the timer 
       "skipping" or "jumping" as it syncs with the server.

