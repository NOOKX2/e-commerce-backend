include .env
export

# สร้าง Connection String สำหรับ migrate tool
DB_URL=postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable

# คำสั่งสำหรับรัน migration "ขึ้น" ทั้งหมดที่ยังไม่ได้รัน
migrateup:
	migrate -path db/migrations -database "$(DB_URL)" -verbose up

# คำสั่งสำหรับรัน migration "ขึ้น" 1 ขั้น
migrateupone:
	migrate -path db/migrations -database "$(DB_URL)" -verbose up 1

# คำสั่งสำหรับรัน migration "ลง" 1 ขั้น
migratedownone:
	migrate -path db/migrations -database "$(DB_URL)" -verbose down 1

