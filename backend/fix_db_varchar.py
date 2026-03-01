import os

filepath = r'd:\\Nagare_Project\\nagare\\backend\\internal\\model\\entities.go'
with open(filepath, 'r', encoding='utf-8') as f:
    content = f.read()

content = content.replace('Value     string    `gorm:"type:varchar(128)" json:"value"`', 'Value     string    `gorm:"type:varchar(2048)" json:"value"`')

with open(filepath, 'w', encoding='utf-8') as f:
    f.write(content)

print("Python script executed.")
