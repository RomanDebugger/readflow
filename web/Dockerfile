FROM node:20-slim AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM python:3.10-slim

COPY --from=golang:1.24-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
COPY --from=builder /app/dist ./dist

EXPOSE 8000
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]