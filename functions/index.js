
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

