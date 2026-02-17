import { initializeApp } from "firebase/app";
import { getAuth, GoogleAuthProvider } from "firebase/auth";
// TODO: Add SDKs for Firebase products that you want to use
// https://firebase.google.com/docs/web/setup#available-libraries

// Your web app's Firebase configuration
// For Firebase JS SDK v7.20.0 and later, measurementId is optional
const firebaseConfig = {
  apiKey: "AIzaSyAv3QynRIoXOdxTusnipK99URUuJg01_As",
  authDomain: "cinema-booking-auth-42d09.firebaseapp.com",
  projectId: "cinema-booking-auth-42d09",
  storageBucket: "cinema-booking-auth-42d09.firebasestorage.app",
  messagingSenderId: "37758616586",
  appId: "1:37758616586:web:6be47d3def06596d848bf3",
  measurementId: "G-EKYJYLQ5P6"
};

// Initialize Firebase
const app = initializeApp(firebaseConfig);


export const auth = getAuth(app);
export const provider = new GoogleAuthProvider();