# Nagare WebSSH: The Command Center (and Why It's Safe)

Imagine your server is a locked room, and you have to go into that room to fix something. Usually, you'd need a special key and a special door (called "SSH"). **Nagare WebSSH** lets you go into that room directly from your web browser, as if you were sitting right in front of the server.

## 1. What is it?
The **WebSSH** is a window inside your Nagare dashboard. You don't need to install anything on your computer. You just click on a server's name and—*BAM!*—you have a "Terminal" (the black screen with white text) to type commands.

## 2. The "Secure Tunnel" Analogy (How it works)
When you type a command in your browser, it's not just "floating" in the open internet.
- **The Messenger (WebSocket)**: Nagare builds a direct, private "phone line" between your browser and the Nagare Brain.
- **The Bodyguard (SSH)**: The Nagare Brain then talks to your server using a very secure, encrypted code. It's like sending a locked box through the mail and only you have the key.

## 3. Why is it secure?
- **No Keys on the Internet**: Your server's password or "private key" never leaves the Nagare server. Your browser never sees it.
- **Encrypted Talk**: Everything you type is scrambled into a secret code (using `AES` and `Chacha20`). Even if a hacker intercepted the message, it would look like gibberish to them.
- **Strict Admittance**: Only people with the right "Permission Level" in Nagare can even see the "Terminal" button.

## 4. Key Features
- **Window Resizing**: You can make the terminal window bigger or smaller, and it "tells" the server to adjust the text so nothing gets cut off.
- **Fast Response**: It uses a technology called `xterm.js` to make the typing feel as fast as a real terminal on your desk.
- **Multi-Tasking**: You can have multiple terminals open at once for different servers, and Nagare keeps track of all of them.

## 5. Summary
Nagare WebSSH is your **"Emergency Hotline"** to your servers. It's built to be as fast as a racing car but as secure as a bank vault.
