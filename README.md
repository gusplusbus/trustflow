# 🚀 Trustflow

Trustflow is a programmable trust platform for open software projects. It allows creators to launch projects backed by verifiable progress, decentralized payouts, and public accountability.

---

## 🧠 Concept

- **Projects** can be created and backed by their own token.
- **Funds** are locked and only released based on verifiable progress.
- **Developers** can apply to take on work, and their contributions are validated via connected GitHub activity (like runners & milestones).
- **Trust** is built through automation, transparency, and proof-of-work.

---

## 🛠️ Tech Stack

| Layer        | Tech                            |
|-------------|---------------------------------|
| Frontend     | HTMX + Go Templates (planned)   |
| Backend      | Go (with Gorilla Mux)           |
| Blockchain   | Node.js (ERC-20, Smart Contracts) |
| Database     | PostgreSQL                       |
| Infra        | Docker, Compose, GitHub Actions  |

---

## 📦 Project Structure



---

## 🚧 Current Status

✅ Docker setup with:
- Go API with static file server
- Health check endpoint
- PostgreSQL container
- Blockchain Node service (stubbed)

✅ Linting with `golangci-lint`  
✅ Base routing structure

---

## 📍 Getting Started

```bash
# Clone the repo
git clone git@github.com:YOUR_USERNAME/trustflow.git
cd trustflow

# Start dev stack
docker compose up --build
