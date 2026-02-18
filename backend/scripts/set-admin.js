const admin = require("firebase-admin");
const serviceAccount = require("./firebase-service-account.json");

admin.initializeApp({
  credential: admin.credential.cert(serviceAccount),
});

// 👉 ใส่ UID ของ user ที่ต้องการให้เป็น admin
const uid = "PUT_USER_UID_HERE";

admin.auth().setCustomUserClaims(uid, { role: "admin" })
  .then(() => {
    console.log("✅ Admin role set successfully");
    process.exit();
  })
  .catch((error) => {
    console.error("❌ Error:", error);
  });