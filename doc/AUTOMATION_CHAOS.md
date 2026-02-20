# Nagare Automation & Chaos: Muscle and Fire Drills

Nagare doesn't just watch; it has the "muscle" to fix things.

## 1. Ansible: The Robot Assistant
Instead of typing commands on 100 servers by hand, Nagare uses **Ansible**.
- **Dynamic Inventory**: Nagare knows exactly which servers are online and tells the robots.
- **Smart Fixes**: If a server is low on disk space, Nagare can automatically send a robot to "Clean Cache."
- **AI Recommendations**: Nagare can ask Gemini: *"Which robot script should I use for this error?"*

## 2. Chaos Engineering: The Fire Drill
How do you know Nagare will work when a real disaster happens? You test it with **Chaos**.
- **Alert Storm**: Nagare can simulate a "Storm" where 100 servers fail at once.
- **Goal**: This tests if your phone gets too many notifications and if the AI can still find the "Root Cause" in the middle of all the noise.

## 3. Triggers: The "If-Then" Rules
You can set simple rules:
- **IF** CPU is $> 90\%$ for 5 minutes...
- **THEN** Run the "Restart-Service" robot.
- **AND** Send a success message to my QQ bot.
