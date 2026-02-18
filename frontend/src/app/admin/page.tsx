"use client";

import { useEffect, useState } from "react";
import { auth } from "@/lib/firebase";
import { useRouter } from "next/navigation";

type Booking = {
  user_id: string;
  seat_id: string;
  movie: string;
  date: string;
  status: string;
};

export default function AdminPage() {
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [movieFilter, setMovieFilter] = useState("");
  const [dateFilter, setDateFilter] = useState("");
  const router = useRouter();

  

  const API_URL = process.env.NEXT_PUBLIC_API_URL;

  const fetchBookings = async () => {
    const user = auth.currentUser;
    if (!user) return;

    const token = await user.getIdToken(true);

    let url = `${API_URL}/admin/bookings?`;
    if (movieFilter) url += `movie=${movieFilter}&`;
    if (dateFilter) url += `date=${dateFilter}`;

    const res = await fetch(url, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    const data = await res.json();
    setBookings(data);
  };

  useEffect(() => {
    const unsubscribe = auth.onAuthStateChanged(async (user) => {
      if (!user){
        router.push("/");
        return;
      } 

      const tokenResult = await user?.getIdTokenResult(true);

      if (tokenResult?.claims.role !== "admin") {
        router.push("/");
        return;
      }

      const idToken = await user.getIdToken(true);

    let url = `${API_URL}/admin/bookings?`;

    if (movieFilter) url += `movie=${movieFilter}&`;
    if (dateFilter) url += `date=${dateFilter}`;

    const res = await fetch(url, {
      headers: {
        Authorization: `Bearer ${idToken}`,
      },
    });

    const data = await res.json();
    
    console.log("📦 Admin bookings:", data);
    setBookings(data);
  });

    return () => unsubscribe();

}, [ ]);

  return (
    <div className="min-h-screen bg-black text-white p-10">
      <h1 className="text-3xl font-bold mb-6">🛠 Admin Dashboard</h1>

      {/* Filters */}
      <div className="flex gap-4 mb-6">
        <select
          value={movieFilter}
          onChange={(e) => setMovieFilter(e.target.value)}
          className="p-2 bg-gray-800 rounded"
        >
          <option value="">All Movies</option>
          <option value="Avengers">Avengers</option>
          <option value="Batman">Batman</option>
        </select>

        <input
          type="date"
          value={dateFilter}
          onChange={(e) => setDateFilter(e.target.value)}
          className="p-2 bg-gray-800 rounded"
        />

        <button
          onClick={fetchBookings}
          className="px-4 py-2 bg-blue-600 rounded"
        >
          Apply Filter
        </button>
      </div>

      {/* Table */}
      <div className="bg-gray-900 rounded-lg overflow-hidden">
        <table className="w-full text-left">
          <thead className="bg-gray-800">
            <tr>
              <th className="p-3">User</th>
              <th className="p-3">Seat</th>
              <th className="p-3">Movie</th>
              <th className="p-3">Date</th>
              <th className="p-3">Status</th>
            </tr>
          </thead>
          <tbody>

            {Array.isArray(bookings) && 
            bookings.map((b, i) => (
              <tr key={i} className="border-b border-gray-700">
                <td className="p-3">{b.user_id}</td>
                <td className="p-3">{b.seat_id}</td>
                <td className="p-3">{b.movie}</td>
                <td className="p-3">{b.date}</td>
                <td className="p-3 text-green-400">{b.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}