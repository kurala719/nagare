# Nagare Reporting System: The Weekly Checkup

Imagine you are a doctor. You don't just want to see a patient when they are sick; you want a weekly report of their health to prevent problems before they start. Nagare's Reporting System does exactly that for your servers.

## 1. What is it?
The Reporting System takes all the "heartbeats" (data) from your servers and turns them into a professional **PDF Report**. You can show this report to your boss or keep it for your own records.

## 2. The "Automated Secretary" (How it works)
You don't have to manually copy and paste numbers into a spreadsheet. Nagare works like an automated secretary:
- **Data Gathering**: Every Sunday at midnight (or whenever you choose), Nagare looks at the last 7 days of data.
- **Smart Analysis**: It calculates things like "Uptime" (how much of the time your server was actually working).
- **PDF Creation**: It uses a tool called `Maroto` to "draw" a beautiful document with your company logo, charts, and tables.

## 3. What's inside the report?
- **The "Executive Summary"**: A quick paragraph written in plain English that tells you if the week was "Good" or "Bad."
- **Visual Charts**: 
  - **Pie Charts**: Show you which servers are the healthiest.
  - **Trend Lines**: Show you when the most problems happened (e.g., "We had a lot of errors on Wednesday afternoon").
- **The "Wall of Shame"**: A list of the top 5 servers that caused the most trouble, so you know exactly what needs fixing.

## 4. Why use it?
- **Proof of Work**: Show your clients that their website was online 99.9% of the time.
- **Trend Spotting**: Notice that a server is getting slower every week *before* it actually crashes.
- **Paper Trail**: If something goes wrong, you have a permanent record of the system's state at that time.

## 5. Fast & Reliable
Generating a big PDF with charts can be hard work. Nagare does this in the "Background" (The Waiting Room). This means you can keep using the app while the secretary is busy typing up your report. Once it's done, you'll get a notification: "Your report is ready!"
