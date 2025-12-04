#!/usr/bin/env python3
"""
Генератор тестовых JWT токенов для chat-service
"""

import jwt
import sys
from datetime import datetime, timedelta

# Секрет из docker-compose.yml (ACCESS_SECRET)
SECRET = "super-secret-access"

def generate_token(role="superadmin", user_id=1, email="test@example.com", 
                   university_id=None, branch_id=None, faculty_id=None):
    """
    Генерирует JWT токен для тестирования
    
    Args:
        role: Роль пользователя (superadmin, curator, operator)
        user_id: ID пользователя
        email: Email пользователя
        university_id: ID университета (опционально)
        branch_id: ID филиала (опционально)
        faculty_id: ID факультета (опционально)
    """
    now = datetime.utcnow()
    exp = now + timedelta(hours=24)  # Токен на 24 часа для тестирования
    
    payload = {
        "sub": str(user_id),
        "email": email,
        "role": role,
        "exp": int(exp.timestamp()),
        "iat": int(now.timestamp())
    }
    
    # Добавляем контекстную информацию
    if university_id is not None:
        payload["university_id"] = university_id
    if branch_id is not None:
        payload["branch_id"] = branch_id
    if faculty_id is not None:
        payload["faculty_id"] = faculty_id
    
    token = jwt.encode(payload, SECRET, algorithm="HS256")
    
    return token, payload

def main():
    print("=== JWT Token Generator ===\n")
    
    # Примеры токенов для разных ролей
    examples = [
        {
            "name": "Superadmin",
            "role": "superadmin",
            "user_id": 1,
            "email": "superadmin@example.com"
        },
        {
            "name": "Curator (University 1)",
            "role": "curator",
            "user_id": 2,
            "email": "curator@example.com",
            "university_id": 1
        },
        {
            "name": "Operator (University 1)",
            "role": "operator",
            "user_id": 3,
            "email": "operator@example.com",
            "university_id": 1
        }
    ]
    
    for i, example in enumerate(examples, 1):
        print(f"{i}. {example['name']}")
        token, payload = generate_token(**{k: v for k, v in example.items() if k != 'name'})
        
        print(f"   Role: {payload['role']}")
        print(f"   User ID: {payload['sub']}")
        print(f"   Email: {payload['email']}")
        if 'university_id' in payload:
            print(f"   University ID: {payload['university_id']}")
        print(f"   Expires: {datetime.fromtimestamp(payload['exp']).strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"\n   Token:\n   {token}\n")
        print(f"   Test command:")
        print(f"   curl -H 'Authorization: Bearer {token}' http://localhost:8082/chats/all\n")
        print("-" * 80 + "\n")

if __name__ == "__main__":
    try:
        import jwt
    except ImportError:
        print("Error: PyJWT library not installed")
        print("Install it with: pip install PyJWT")
        sys.exit(1)
    
    main()
