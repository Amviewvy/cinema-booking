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

type AuditLog = {
  Event: string;
  UserID: string;
  ShowID: string;
  SeatID: string;
  Message: string;
  timestamp: string;
};

export default function AdminPage() {
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [movieFilter, setMovieFilter] = useState("");
  const [dateFilter, setDateFilter] = useState("");
  const router = useRouter();
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [showLogs, setShowLogs] = useState(false);


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

    if (!Array.isArray(data)) {
      console.error("Invaild bookings response: ", data);
      setBookings([]);
      return;
    }
    

    setBookings(data);
  };

  const fetchLogs = async () => {
    const user = auth.currentUser;
    if (!user) return;

    const token = await user.getIdToken(true);

    const res = await fetch(`${API_URL}/admin/logs`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    
    if (!res.ok) {
      console.error("API Error: ", res.status);
      return;
    }

    const data = await res.json();

    if (!Array.isArray(data)) {
      console.error("Logs is not array: ", data);
      return;
    }

    setLogs(data);
    setShowLogs(true);

    console.log("Logs: ", data);
    
    //console.log("Token: ", token);
    //console.log("Calling:", `${API_URL}/admin/logs`);
  };
  
  
  //console.log("ADMIN API_URL:", process.env.NEXT_PUBLIC_API_URL);

  const handleLogout = async () => {
    await auth.signOut();
    router.push("/");
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

    // if (tokenResult?.claims.role === "admin") {
    //   router.push("/admin");
    //   return;
    // }

    await fetchBookings();
    await fetchLogs();
    });

    return () => unsubscribe();

}, [ ]);

useEffect(() => {
  if (!API_URL) {
    console.error("API_URL is undefined");
    return;
  }

  const wsUrl = API_URL.replace("http", "ws") + "/ws";
  const ws = new WebSocket(wsUrl);

  if (!auth.currentUser) return;

  ws.onopen = () => {
    console.log("Admin ws connected ");
  }

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);

    if (data.event === "seat_booked" || 
      data.event === "seat_locked" || 
      data.event === "seat_released") {
      fetchBookings();
      fetchLogs();
    }
  };

  return () => ws.close();
}, [API_URL]);

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
      <div className="flex-1 text-right">
        <button
          onClick={handleLogout}
          className="text-red-400  hover:underline"
        >
          Logout
        </button>
      </div>
        

        
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
            {showLogs && (
        <div className="mt-10">
          <h2 className="text-xl mb-4">📜 Audit Logs</h2>
          <table className="w-full text-left bg-gray-900 rounded">
            <thead className="bg-gray-800">
              <tr>
                <th className="p-3">Event</th>
                <th className="p-3">User</th>
                <th className="p-3">Seat</th>
                <th className="p-3">Message</th>
              </tr>
            </thead>
            <tbody>
              {logs.map((log, i) => (
                <tr key={i} className="border-b border-gray-700">
                  <td className="p-3 text-yellow-400">{log.Event}</td>
                  <td className="p-3">{log.UserID}</td>
                  <td className="p-3">{log.SeatID}</td>
                  <td className="p-3">{log.Message}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}