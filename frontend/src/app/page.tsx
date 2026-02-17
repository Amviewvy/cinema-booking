"use client";

import { useEffect, useState } from "react";
import { signInWithPopup, signOut, onAuthStateChanged } from "firebase/auth";
import { auth, provider } from "../lib/firebase";

type Seat = {
  seat_id: string;
  status: string;
};

export default function Home() {
  const [user, setUser] = useState<any>(null);
  const [token, setToken] = useState<string | null>(null);
  const [seats, setSeats] = useState<Seat[]>([]);
  const [selectedSeat, setSelectedSeat] = useState<string | null>(null);
  const [isLocked, setIsLocked] = useState(false);

  const API_URL = process.env.NEXT_PUBLIC_API_URL;

  useEffect(() => {
    const unsubscribe = onAuthStateChanged(auth, async (user) => {
      if (user) {
        const idToken = await user.getIdToken();
        setUser(user);
        setToken(idToken);
        fetchSeats();
      }
    });

    return () => unsubscribe();
  }, []);

  const fetchSeats = async () => {
    const res = await fetch(`${API_URL}/seats?show_id=show1`);
    const data = await res.json();
    setSeats(data);

    if (selectedSeat) {
      const currentSeat = data.find((s: Seat) => s.seat_id === selectedSeat);
      if (currentSeat?.status !== "LOCKED") {
        setIsLocked(false);
        setSelectedSeat(null);
      }
    }
  };

  const handleLogin = async () => {
    await signInWithPopup(auth, provider);
  };

  const handleLogout = async () => {
    await signOut(auth);
    setUser(null);
    setSeats([]);
  };

  const handleBooking = async () => {
    if (!token || !selectedSeat) return;

    const res = await fetch(`${API_URL}/seats/lock`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        show_id: "show1",
        seat_id: selectedSeat,
      }),
    });

    if(res.ok) {
      setIsLocked(true);
    }else { 
      setIsLocked(false);
    }

    await fetchSeats();
    
  };

  const handlePayment = async () => {
    if (!token || !selectedSeat) return;

    const res = await fetch(`${API_URL}/payment/success`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        show_id: "show1",
        seat_id: selectedSeat,
      }),
    });

    const data = await res.json();
    if (data.success) {
      alert("Payment successful! Your seat is booked.");
    } else {
      alert("Payment failed. Please try again.");
    }
    console.log(data);

    await fetchSeats();

  };

  const getSeatColor = (status: string) => {
    if (status === "AVAILABLE") return "bg-green-500";
    if (status === "LOCKED") return "bg-yellow-500";
    if (status === "BOOKED") return "bg-red-500";
    return "bg-gray-400";
  };

  return (
    <div className="min-h-screen bg-black text-white p-8">
      <h1 className="text-3xl font-bold mb-6 text-center">
        🎬 Cinema Ticket Booking
      </h1>

      {!user ? (
        <div className="flex justify-center">
          <button
            onClick={handleLogin}
            className="px-6 py-3 bg-blue-500 rounded-lg"
          >
            Login with Google
          </button>
        </div>
      ) : (
        <>
          <div className="flex justify-between mb-6">
            <div>Welcome, {user.displayName}</div>
            <button onClick={handleLogout} className="text-red-400">
              Logout
            </button>
          </div>

          <h2 className="text-xl mb-4">🎥 Show: Avengers - 7PM</h2>

          <div className="grid grid-cols-5 gap-4 mb-6">
            {seats.map((seat) => (
              <button
                key={seat.seat_id}
                onClick={() =>
                  seat.status === "AVAILABLE" &&
                  setSelectedSeat(seat.seat_id)
                }
                className={`p-4 rounded ${getSeatColor(
                  seat.status
                )} ${
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
                Confirm Booking
              </button>
            </div>
          )}
          {isLocked ? (
            <button
              onClick={handlePayment}
              className="px-6 py-3 bg-green-600 rounded-lg"
            >
              Pay Now 💳
            </button>
          ) : (
            <button disabled className="px-6 py-3 bg-gray-500 rounded-lg cursor-not-allowed">
              Lock seat first
            </button>
          )}

          
        </>
      )}
    </div>
  );
}