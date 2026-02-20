"use client";

import { use, useEffect, useState } from "react";
import { signInWithPopup, signOut, onAuthStateChanged } from "firebase/auth";
import { auth, provider } from "../lib/firebase";
import { useRouter } from "next/navigation";

type Seat = {
  seat_id: string;
  status: string;
};

export default function Home() {
  const [user, setUser] = useState<any>(null);
  const [token, setToken ] = useState<string | null>(null);
  const [seats, setSeats] = useState<Seat[]>([]);
  const [selectedSeat, setSelectedSeat] = useState<string | null>(null);
  const [isLocked, setIsLocked] = useState(false);
  const [countdown, setCountdown] = useState<number>(0);
  const [role, setRole] = useState<string | null>(null);
  const router = useRouter();
  const [shows, setShows] = useState<any[]>([]);
  const [selectedShow, setSelectedShow] = useState("show1");
 
  const API_URL = process.env.NEXT_PUBLIC_API_URL;

  const currentShow = shows.find((s) => s.id === selectedShow);


useEffect(() => {
  if (!API_URL) return; 

  const fetchShows = async () => {
    try {
    const res = await fetch(`${API_URL}/shows`);

    if (!res.ok) {
      console.error("Failed to fetch shows: ");
      setShows([]);
      return;
    }


    const data = await res.json();
    setShows(Array.isArray(data) ? data : []);

    
  } catch (error) {
    console.error("Failed to fetch shows: ", error);
    setShows([]);
  }
  };
  

  fetchShows();
  
}, [API_URL]);


const fetchSeats = async () => {
  const user = auth.currentUser;

    const res = await fetch(`${API_URL}/seats?show_id=${selectedShow}`);
    //const res = await fetch(`${API_URL}/seats?show_id=${showId}`);
    const data = await res.json();

    const seatArray = Array.isArray(data) ? data : [];
    setSeats(seatArray);

    if (!user) return;

    const myLockedSeat = seatArray.find(
    (s: any) =>
      s.status === "LOCKED" &&
      s.locked_by === user.uid
  );

  if (myLockedSeat) {
    setSelectedSeat(myLockedSeat.seat_id);
    setIsLocked(true);

    if (myLockedSeat.lock_expire) {
      const expireTime = new Date(myLockedSeat.lock_expire).getTime();
      const remaining = Math.floor((expireTime - Date.now()) / 1000);

      if (remaining > 0) {
        setCountdown(remaining);
      } else {
        setIsLocked(false);
      }
    }
  }
  };


useEffect(() => {
  if (selectedShow) {
    fetchSeats();
    setSelectedSeat(null);
    setIsLocked(false);
  }
}, [selectedShow]);


  useEffect(() => {
    const unsubscribe = onAuthStateChanged(auth, async (user) => {
      if (user) {
    
        const idToken = await user.getIdToken(true);
        const tokenResult = await user.getIdTokenResult(true);

        const userRole = (tokenResult.claims.role as string) || "user";

        if (userRole === "admin") {
          router.push("/admin");
          return;
        }

      
        console.log("📋 Token Claims:", tokenResult.claims);
        console.log("👤 User Role:", userRole);

        setRole(userRole);
        setUser(user);
        setToken(idToken);
        fetchSeats();
      }
    });

    return () => unsubscribe();
    
  }, []);

  
  
useEffect(() => {
  if (!API_URL) return;

  const wsUrl = API_URL.replace("http", "ws") + "/ws";
  const ws = new WebSocket(wsUrl);

  ws.onopen = () => {
    console.log("✅ WebSocket connected");
  };

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log("📡 WS message:", data);

    if (data.event?.startsWith("seat_")) {
      fetchSeats();
    }
  };

  ws.onclose = () => {
    console.log("❌ WebSocket disconnected");
  };

  return () => {
    ws.close();
  };
}, [API_URL]);

useEffect(() => {
  if (!isLocked || countdown <= 0) return;

  const timer = setInterval(() => {
    setCountdown((prev) => prev - 1);
  }, 1000);

  return () => clearInterval(timer);
}, [isLocked, countdown]);

useEffect(() => {
  if (countdown === 0 && isLocked) {
    setIsLocked(false);
    setSelectedSeat(null);
  }
}, [countdown]);

  

  const handleLogin = async () => {
    await signInWithPopup(auth, provider);
  };

  const handleLogout = async () => {
    await signOut(auth);
    setUser(null);
    setSeats([]);
  };

  const handleBooking = async () => {
    const user = auth.currentUser;
    if (!user) return;

    const token = await user.getIdToken(true);

    if (!token || !selectedSeat) return;

    const res = await fetch(`${API_URL}/seats/lock`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        //show_id: "show1",
        show_id: selectedShow,
        seat_id: selectedSeat,
      }),
    });

    if(res.ok || res.status === 409) {
      setIsLocked(true);
      setCountdown(60); 
    }else { 
      setIsLocked(false);
    }

    await fetchSeats();
    
  };

  const handlePayment = async () => {
    const user = auth.currentUser;
    if (!user) {
      alert("Please login to proceed with payment.");
      return;
    }

    const token = await user.getIdToken(true);

    if (!token || !selectedSeat) return;

    const res = await fetch(`${API_URL}/payment/success`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        //show_id: "show1",
        show_id: selectedShow,
        seat_id: selectedSeat,
      }),
    });

    const data = await res.json();

    if (!res.ok) {
      alert("Payment failed: " + data.error );
      return;
    }
   

    alert("🎉 Payment successful! Your seat is booked.");


    await fetchSeats();
    setIsLocked(false);
    setSelectedSeat(null);

  };

  const getSeatColor = (status: string) => {
    if (status === "AVAILABLE") return "bg-green-500";
    if (status === "LOCKED") return "bg-yellow-500";
    if (status === "BOOKED") return "bg-red-500";
    return "bg-gray-400";
  };

  return (
  <div className="min-h-screen bg-black text-white p-8">
    {!user ? (
      <div className="flex justify-center">
        <button
          onClick={handleLogin}
          className="px-6 py-3 bg-blue-500 rounded-lg"
        >
          Login with Google
        </button>
      </div>
    )  : (
      <>
        <h1 className="text-3xl font-bold mb-6 text-center">
          🎬 Cinema Ticket Booking
        </h1>

        <div className="flex justify-between mb-6">
          <div>Welcome, {user.displayName}</div>
          <button onClick={handleLogout} className="text-red-400 hover:underline">
            Logout
          </button>
        </div>

        <select
        value={selectedShow}
        onChange={(e) => setSelectedShow(e.target.value)}
        className="p-2 bg-gray-800 rounded"
      >
        {Array.isArray(shows) && shows.map((show) => (
          <option key={show.id} value={show.id}>
            {show.name}
          </option>
        ))}
      </select>

        <h2 className="text-xl mb-4">🎥 Show:  {currentShow?.name}</h2>

    
      

        <div className="grid grid-cols-5 gap-4 mb-6">
          {Array.isArray(seats) && seats.map((seat) => (
            <button
              key={`${selectedShow}-${seat.seat_id}`}
              onClick={() =>
                seat.status === "AVAILABLE" &&
                setSelectedSeat(seat.seat_id)
              }
              className={`p-4 rounded ${getSeatColor(seat.status)} ${
                selectedSeat === seat.seat_id
                  ? "border-4 border-white"
                  : ""
              }`}
            >
              {seat.seat_id}
            </button>
          ))}
        </div>

        {selectedSeat && (
          <div className="text-center">
            <p className="mb-4">
              Selected Seat: <b>{selectedSeat}</b>
            </p>
            <button
              onClick={handleBooking}
              className="px-6 py-3 bg-purple-600 rounded-lg"
            >
              LOCK SEAT
            </button>
          </div>
        )}

        {isLocked ? (
          <button
            onClick={handlePayment}
            className="px-6 py-3 bg-green-600 rounded-lg"
          >
            Pay Now 💳 ({countdown}s)
          </button>
        ) : (
          <button
            disabled
            className="px-6 py-3 bg-gray-500 rounded-lg cursor-not-allowed"
          >
            Choose your seat
          </button>
        )}

        
      </>
    )}
  </div>
);
}