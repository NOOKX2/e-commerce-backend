FROM golang:1.25-alpine AS builder

# ตั้งค่า Working Directory ภายใน builder container
WORKDIR /app

# 1. Copy go.mod และ go.sum ก่อน
#    เพื่อใช้ประโยชน์จาก Docker layer caching
#    ถ้าไฟล์ 2 นี้ไม่เปลี่ยน Docker จะไม่ดาวน์โหลด dependencies ใหม่

COPY go.mod ./
COPY go.sum ./
RUN go mod download

# 2. Copy source code ทั้งหมด
COPY . .



# 3. Build แอปพลิเคชัน
#    - CGO_ENABLED=0 เพื่อสร้าง static binary ที่ไม่ขึ้นกับ C libraries
#    - -o /bin/app คือ output ไปที่ /bin/app
#    - ./cmd/api คือ path ไปยัง package main ของคุณ
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/app ./cmd/api

# --- Stage 2: The Final Image ---
# ใช้ base image ที่เล็กที่สุดเท่าที่เป็นไปได้
# scratch คือ image ว่างเปล่า, alpine เล็กและมี shell ให้ debug ได้
FROM alpine:latest

# (ถ้าจำเป็น) ติดตั้ง CA certificates สำหรับการเชื่อมต่อ HTTPS/TLS
RUN apk --no-cache add ca-certificates

# ตั้งค่า Working Directory
WORKDIR /root/

# Copy เฉพาะไฟล์ binary ที่คอมไพล์เสร็จแล้วจาก Stage "builder"
COPY --from=builder /bin/app .

# (ถ้ามี) Copy ไฟล์ configs เข้ามาด้วย
COPY configs/ ./configs/

# บอกให้รู้ว่า container จะ listen ที่พอร์ตอะไร (ต้องตรงกับในโค้ด)
EXPOSE 8080

# คำสั่งสำหรับรันแอปพลิเคชันเมื่อ container เริ่มทำงาน
CMD ["./app"]