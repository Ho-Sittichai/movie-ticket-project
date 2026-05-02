# Movie Ticket Booking System - ระบบจองตั๋วหนังแบบ Real-time 🎬🍿

**Movie Ticket Booking System** คือโปรเจกต์ที่ให้ผู้ใช้งานสามารถซื้อตั๋วหนังได้แบบง่ายๆ โดยเน้นความถูกต้องของข้อมูล (Data Integrity) และประสิทธิภาพที่ต้องรองรับคนจำนวนมากได้แบบไม่ค้าง(Concurrency)

---

## 🏗️ 1. Architecture (ภาพรวมระบบ)

ระบบนี้ผมออกแบบมาให้เป็นแบบ **Decoupled Architecture** เพื่อแยกส่วนการทำงานให้ชัดเจนและรองรับการขยายตัว (Scalability) ในอนาคตครับ

```mermaid
graph TD
    User((👤 Users)) -->|HTTP/HTTPS| FE[🌐 Vue.js Frontend]
    FE -->|WebSocket/REST| BE[🚀 Go Backend Services]

    subgraph "Data & Consistency Layer"
        BE -->|SETNX| Redis[(⚡ Redis: Distributed Lock)]
        BE -->|Store| Mongo[(🍃 MongoDB: Persistent DB)]
    end

    subgraph "Event-Driven Background"
        BE -->|Produce| Kafka{{"📦 Apache Kafka"}}
        Kafka -->|Consume| Workers[📧 Notification/Audit Services]
    end
```

---

## 🛠️ 2. Tech Stack Overview

### Backend (Golang)

- **Framework**: [Gin Gonic](https://github.com/gin-gonic/gin) (Web Framework ที่เน้นความเร็วและประสิทธิภาพสูง)
- **Database**: [MongoDB](https://www.mongodb.com/) (เก็บข้อมูลหนังและรอบฉายในรูปแบบ Document ที่ยืดหยุ่น)
- **Cache / Lock**: [Redis](https://redis.io/) (ใช้ทำ **Distributed Lock** เท่านั้น ไม่ใช่แค่ Cache)
- **Messaging**: [Apache Kafka](https://kafka.apache.org/) (ใช้ทำ Event-Driven Architecture สำหรับระบบแจ้งเตือน)
- **Real-time**: [Gorilla WebSocket](https://github.com/gorilla/websocket) (ใช้ส่งสถานะที่นั่งแบบสดๆ หาผู้ใช้ทุกคน)
- **Auth**: [Google OAuth 2.0](https://developers.google.com/identity/protocols/oauth2) (ระบบยืนยันตัวตนผ่าน Google Account)

### Frontend (Vue.js)

- **Framework**: [Vue.js](https://vuejs.org/) Vue 3 + Vite (สำหรับสร้างหน้าเว็บ และ คอมไพล์หน้าเว็บให้ใช้งานได้อย่างรวดเร็ว)
- **Styling**: [TailwindCSS](https://tailwindcss.com/) / Vanilla CSS (สำหรับตกแต่ง UI ต่างๆในแอปให้มีความสวยงาม)
- **State Management**: [Pinia](https://pinia.vuejs.org/) (จัดการ State ภายในแอป)

### Deployment (Infrastructure)

- **Containerization**: [Docker](https://www.docker.com/) + Docker Compose (จำลองสภาพแวดล้อม Production ให้รันได้ทุกที่)
- **Monitoring**: [Kafdrop](https://github.com/obsidiandynamics/kafdrop) (UI สำหรับดู Kafka Topic/Message)

---

## 🔄 3. Booking Flow (ขั้นตอนการทำงาน)

1.  **Sign In (เข้าสู่ระบบ)**: ผู้ใช้เข้าสู่ระบบผ่าน Google Account **(เพื่อความใช้งานง่าย Mock คนเข้าคนแรกเป็น Admin ที่เหลือเป็น User)**
2.  **Selection (เลือกหนัง)**: ผู้ใช้เลือก `Movie` และ `Showtime` ที่ต้องการ
3.  **Seat Locking (ล็อคที่นั่ง Real-time)**:
    - ทันทีที่คลิกที่นั่ง **Frontend** จะยิง `POST /api/lock`
    - **Backend** จะสร้าง **Distributed Lock** ใน Redis โดยอ้างอิงจาก `MovieID` + `StartTime`
    - **WebSocket** กระจายสถานะ `LOCKED` ไปยังผู้ใช้คนอื่นๆ ที่ดูรอบเดียวกันทันที (คนอื่นจะเห็นเป็นสีแดง)
4.  **Payment Initiation (เข้าสู่การจ่ายเงิน)**:
    - เมื่อกด "จอง/ชำระเงิน" ระบบจะต่ออายุ Lock (Extend TTL) ให้เป็น 5 นาที
    - สร้าง **Payment Lock** เพื่อกันไม่ให้ผู้ใช้ไปกดเปลี่ยนที่นั่งระหว่างจ่ายเงิน
5.  **Confirmation (ยืนยันผล)**:
    - เมื่อชำระเงินสำเร็จ (Mock) **Backend** จะอัปเดตสถานะเป็น `BOOKED` ลง MongoDB (Atomic Update)
    - ปลดล็อค Redis (Release Lock)
    - **WebSocket** แจ้งทุกคนว่าที่นั่งนี้ "ขายแล้ว" (`BOOKED`)
6.  **Audit & Notification (ทำงานเบื้องหลัง)**:
    - Backend ส่ง Event `BOOKING_SUCCESS` เข้าไปใน **Kafka**
    - **Email Service** (Consumer) ที่รออยู่ จะหยิบ Event ไปส่งอีเมลยืนยัน **(Mock ในหลังบ้าน)**
    - **Audit Log** บันทึกประวัติการจองเก็บไว้ตรวจสอบ

---

## 🔐 4. Redis Lock Strategy (หัวใจสำคัญ)

เพื่อแก้ปัญหา **"Double Booking"** (แย่งที่นั่งกัน)

- **Uniqueness**: รับประกันความไม่ซ้ำกัน 100% ด้วยการใช้ ID หนัง + เวลาฉายจริง + ID ที่นั่ง
- **Atomicity (SETNX)**: ใช้คำสั่ง `SETNX` ของ Redis ซึ่งเป็น Atomic Operation ทำให้มั่นใจได้ว่าจะมี **"ผู้ที่ได้ล็อคที่นั่งคนเดียว"** ที่ได้ล็อคที่นั่งไปครอง แม้จะมีการกดจองเข้ามาพร้อมกันในเสี้ยววินาทีก็ตาม
- **TTL (Auto-Expire)**: ถ้าคนจองหายไปเฉยๆ ล็อคจะหลุดเองภายใน 5 นาที (ป้องกัน Deadlock ที่นั่งค้าง)
- **Ownership**: เฉพาะคนที่ล็อคเท่านั้นที่มีสิทธิ์ปลดล็อค หรือจองที่นั่งนั้นต่อได้

---

## 📨 5. Message Queue Strategy (Kafka)

- **Topic**: `booking_events`
- **Events**:
  - `BOOKING_SUCCESS`: ยิงเพื่อส่งเมล
  - `SYSTEM_ERROR` / `AUDIT_LOG`: เก็บ Log ไว้ดูย้อนหลัง
- **ข้อดี**: ถ้าเซิร์ฟเวอร์ส่งเมลล่ม หรือช้า ผู้ใช้ก็ยังจองตั๋วได้สำเร็จตามปกติ ไม่ต้องรอนาน ระบบจะ Retry ส่งเมลให้เองทีหลัง

---

## 🚀 6. How to Run (วิธีรันระบบ)

**สิ่งที่ต้องมี**: Docker & Docker Compose

1.  **Clone & Start**:

    ```bash
    git clone https://github.com/Ho-Sittichai/movie-ticket-project.git
    cd movie-ticket-project
    ```

2.  **Setup Environment Variables**:
    สร้างไฟล์ `.env` ในโฟลเดอร์ `backend` และกำหนดค่า Google OAuth ดังนี้:
    _(Create `.env` in `backend` folder and add these variables)_

    ```env
    GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/google/callback
    GOOGLE_CLIENT_ID=your_google_client_id
    GOOGLE_CLIENT_SECRET=your_google_client_secret
    ```

3.  **Run Application**:

    ```bash
    docker-compose up --build
    ```

4.  **Access**:
    - Frontend (หน้าเว็บ): `http://localhost:5173`
    - Backend API: `http://localhost:8080`
    - Kafka UI (ดู Event): `http://localhost:9000`

---

## ⚖️ 7. Assumptions & Trade-offs (ข้อควรคำนึง)

- **Consistency vs. Availability**: ในระบบการเงิน/การจอง เราเน้น **Strong Consistency** (ข้อมูลต้องเป๊ะ) มากกว่าความเร็วเพียงอย่างเดียว เราจึงใช้ Redis Lock + Mongo Atomic Update
- **Seat Mapping**: ในเดโมนี้สมมติว่าเป็นผังที่นั่งแบบ Fixed เพื่อความง่ายต่อการเข้าใจ (ของจริงอาจต้องเป็น Dynamic Grid)
- **Kafka Overhead**: สำหรับโปรเจกต์ขนาดเล็ก Kafka อาจดูใหญ่เกินความจำเป็น แต่เราใส่มาเพื่อโชว์โครงสร้างแบบ **Enterprise Scalability** และ **Event-Driven Pattern**

---

## 🎥 8. Example Page & Demo (ตัวอย่างหน้าเว็บและวิดีโอการใช้งาน)

### ตัวอย่างหน้าเว็บ (Example Page)
![Example Page](./picture/image.png)

### วิดีโอสาธิตการใช้งาน (Demo Video)
📺 [คลิกเพื่อดูวิดีโอการเล่นเว็บนี้เบื้องต้น](https://drive.google.com/file/d/1nkata15V6sRf88wOAEBftjIPCwlbQtjD/view?usp=drive_link)

