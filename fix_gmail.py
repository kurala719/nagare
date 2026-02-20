import os

filename = "backend/internal/repository/media/gmail.go"

with open(filename, 'r', encoding='utf-8') as f:
    content = f.read()

# Replace actual literal newline in the string with \r\n
# The error says "newline in string" meaning there's a literal newline.
# We will use simple string replace
content = content.replace('From: %s\n"+\n\t\t"To: %s', 'From: %s\\r\\n"+\n\t\t"To: %s')

# Actually, it's easier to find the exact blocks and replace them.
# I will use regex

import re
fixed_content = re.sub(r'fmt\.Sprintf\("From:.*?--%s--",', r'''fmt.Sprintf("From: %s\\r\\n" +
		"To: %s\\r\\n" +
		"Subject: %s\\r\\n" +
		"MIME-Version: 1.0\\r\\n" +
		"Content-Type: multipart/alternative; boundary=%s\\r\\n" +
		"\\r\\n" +
		"--%s\\r\\n" +
		"Content-Type: text/plain; charset=\\"UTF-8\\"\\r\\n" +
		"Content-Transfer-Encoding: base64\\r\\n" +
		"\\r\\n" +
		"%s\\r\\n" +
		"--%s--",''', content, flags=re.DOTALL)

with open(filename, 'w', encoding='utf-8') as f:
    f.write(fixed_content)
