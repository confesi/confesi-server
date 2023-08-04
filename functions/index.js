
// See a full list of supported triggers at https://firebase.google.com/docs/functions

const {onRequest} = require("firebase-functions/v2/https");
const logger = require("firebase-functions/logger");
const functions = require("firebase-functions");
const admin = require("firebase-admin");

admin.initializeApp(); // Initialize Firebase Admin SDK

// Test auth function trigger
exports.userSignUp = functions.auth.user().onCreate((user) => {
    console.log(user);
    const userId = user.uid;
    const email = user.email;

    console.log(`New user signed up: ${email}, UID: ${userId}`);
    return Promise.resolve();
});

exports.clearInactiveUsers = functions.pubsub
  .schedule("every 24 hours") // every 24 hours
  .timeZone("UTC") 
  .onRun(async (context) => {
    try {
      const now = Date.now();
      const twoDaysAgo = now - 2 * 24 * 60 * 60 * 1000; // two days in ms

      // fetch all accounts
      const userAccounts = await admin.auth().listUsers();

      const deletions = [];

        // find all users that have not verified their email and are older than two days
      userAccounts.users.forEach((user) => {
        if (!user.emailVerified && user.metadata.creationTime < twoDaysAgo) {
          deletions.push(admin.auth().deleteUser(user.uid));
        }
      });

      // Execute all user deletions concurrently
      await Promise.all(deletions);

      console.log("[JOB] non-email-verified users deleted successfully");
      return Promise.resolve();
    } catch (error) {
      console.error("error while deleting non-email-verified users:", error);
      throw new functions.https.HttpsError("internal", "error while deleting non-email-verified users");
    }
  });

