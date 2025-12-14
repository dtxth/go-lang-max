#!/usr/bin/env python3
"""
Генератор тестовых JWT токенов для auth-service
"""

import jwt
import uuid
from datetime import datetime, timedelta

# Секреты из docker-compose.yml
ACCESS_SECRET = "super-secret-access"
REFRESH_SECRET = "super-secret-refresh"

def generate_tokens():
    # Используем время контейнера (UTC) и добавляем запас времени
    now = datetime.utcnow()
    print(f"Current time: {now.strftime('%Y-%m-%d %H:%M:%S')} UTC")
    
    # Access token (15 минут)
    access_exp = now + timedelta(hours=1)  # Увеличиваем для тестирования
    access_payload = {
        "sub": "1",
        "email": "operator@example.com",
        "role": "operator",
        "exp": int(access_exp.timestamp()),
        "iat": int(now.timestamp()),
        "university_id": 1
    }
    access_token = jwt.encode(access_payload, ACCESS_SECRET, algorithm="HS256")
    
    # Refresh token (7 дней)
    refresh_exp = now + timedelta(days=7)
    refresh_payload = {
        "sub": "1",
        "email": "operator@example.com",
        "role": "operator",
        "exp": int(refresh_exp.timestamp()),
        "iat": int(now.timestamp()),
        "jti": str(uuid.uuid4()),
        "university_id": 1
    }
    refresh_token = jwt.encode(refresh_payload, REFRESH_SECRET, algorithm="HS256")
    
    print("\n=== JWT Token Generator ===\n")
    print("Operator (реальный пользователь из БД)")
    print(f"Role: {access_payload['role']}")
    print(f"User ID: {access_payload['sub']}")
    print(f"Email: {access_payload['email']}")
    print(f"University ID: {access_payload['university_id']}")
    
    print(f"\nACCESS TOKEN (expires: {datetime.fromtimestamp(access_payload['exp']).strftime('%Y-%m-%d %H:%M:%S')} UTC):")
    print(f"{access_token}\n")
    
    print(f"REFRESH TOKEN (expires: {datetime.fromtimestamp(refresh_payload['exp']).strftime('%Y-%m-%d %H:%M:%S')} UTC):")
    print(f"{refresh_token}\n")
    
    print("Test commands:")
    print(f"# Тест с access токеном:")
    print(f"curl -H 'Authorization: Bearer {access_token}' http://localhost:8082/chats/all")
    print(f"\n# ВНИМАНИЕ: Сгенерированный refresh токен НЕ БУДЕТ работать,")
    print(f"# так как он не сохранен в базе данных!")
    print(f"# Используйте логин для получения действующих токенов:")
    print(f"curl -X POST -H 'Content-Type: application/json' \\")
    print(f"  -d '{{\"email\": \"operator@example.com\", \"password\": \"password\"}}' \\")
    print(f"  http://localhost:8080/login")
    print(f"\n# Затем используйте полученный refresh_token для обновления:")
    print(f"# curl -X POST -H 'Content-Type: application/json' \\")
    print(f"#   -d '{{\"refresh_token\": \"<полученный_refresh_token>\"}}' \\")
    print(f"#   http://localhost:8080/refresh")

if __name__ == "__main__":
    try:
        import jwt
    except ImportError:
        print("Error: PyJWT library not installed")
        print("Install it with: pip install PyJWT")
        exit(1)
    
    generate_tokens()