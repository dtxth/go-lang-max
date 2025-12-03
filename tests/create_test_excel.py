#!/usr/bin/env python3
"""Создание тестового Excel файла для импорта"""

try:
    from openpyxl import Workbook
except ImportError:
    print("❌ Установите openpyxl: pip install openpyxl")
    exit(1)

# Создаем workbook
wb = Workbook()
ws = wb.active

# Header row
headers = [
    "Phone1", "MaxID", "INN_Ref", "FOIV", "OrgName", "Branch",
    "INN", "KPP", "Faculty", "Course", "Group", "ChatName",
    "Phone2", "FileName", "ChatID", "Link", "AddUser", "AddAdmin",
]

ws.append(headers)

# Data row
data_row = [
    "79884753064", "496728250", "105014177", "Минобрнауки России",
    "МГТУ", "Головной филиал", "105014177", "10501001",
    "Политехнический колледж МГТУ", "2", "Колледж ИП-22",
    "Колледж ИП-22 (2024 ОФО МГТУ", "79884753064", "file.xlsx",
    "-69257108032233", "https://max.ru/join/test", "ИСТИНА", "ИСТИНА",
]

ws.append(data_row)

# Сохраняем
wb.save("/tmp/test_import.xlsx")
print("✅ Test Excel file created: /tmp/test_import.xlsx")
