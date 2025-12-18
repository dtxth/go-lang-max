# Тесты для методов обновления названий элементов структуры

## Обзор

Создан comprehensive набор тестов для новых методов обновления названий элементов иерархической структуры университета. Тесты покрывают все уровни архитектуры: HTTP handlers, use cases и включают property-based и benchmark тестирование.

## Структура тестов

### 1. HTTP Handler Tests
**Файл**: `internal/infrastructure/http/handler_name_update_test.go`

#### Покрытие:
- ✅ **Успешные сценарии** - обновление названий всех типов элементов
- ✅ **Валидация входных данных** - некорректные ID, пустые названия
- ✅ **Обработка ошибок** - несуществующие записи, внутренние ошибки
- ✅ **Маршрутизация** - правильность HTTP routing

#### Тестируемые endpoints:
- `PUT /universities/{id}/name`
- `PUT /branches/{id}/name`
- `PUT /faculties/{id}/name`
- `PUT /groups/{id}/name`

#### Результаты:
```
=== RUN   TestUpdateUniversityName_Success
--- PASS: TestUpdateUniversityName_Success (0.00s)
=== RUN   TestUpdateBranchName_Success
--- PASS: TestUpdateBranchName_Success (0.00s)
=== RUN   TestUpdateFacultyName_Success
--- PASS: TestUpdateFacultyName_Success (0.00s)
=== RUN   TestUpdateGroupName_Success
--- PASS: TestUpdateGroupName_Success (0.00s)
```

### 2. Use Case Tests
**Файл**: `internal/usecase/structure_service_name_update_test.go`

#### Покрытие:
- ✅ **Бизнес-логика** - корректность обновления названий
- ✅ **Сохранение данных** - неизменность других полей
- ✅ **Обработка ошибок** - правильная пропагация ошибок репозитория
- ✅ **Edge cases** - пустые названия, отсутствующие связи

#### Тестируемые методы:
- `UpdateUniversityName(id int64, name string) error`
- `UpdateBranchName(id int64, name string) error`
- `UpdateFacultyName(id int64, name string) error`
- `UpdateGroupName(id int64, name string) error`
- `GetBranchByID(id int64) (*Branch, error)`
- `GetFacultyByID(id int64) (*Faculty, error)`

#### Сценарии тестирования:
1. **Успешное обновление** - проверка корректности изменения названия
2. **Запись не найдена** - обработка ошибок `ErrNotFound`
3. **Ошибка обновления** - обработка ошибок репозитория
4. **Сохранение полей** - неизменность других атрибутов
5. **Особые случаи** - группы без чатов, факультеты без филиалов

### 3. Property-Based Tests
**Файл**: `internal/usecase/structure_service_name_update_properties_test.go`

#### Использует библиотеку: `github.com/leanovate/gopter`

#### Тестируемые свойства:
- ✅ **Сохранение полей** - все поля кроме названия остаются неизменными
- ✅ **Unicode поддержка** - корректная обработка кириллицы и эмодзи
- ✅ **Пустые строки** - обработка пробелов и пустых названий
- ✅ **Идемпотентность** - повторное обновление тем же значением
- ✅ **Пропагация ошибок** - корректная передача ошибок репозитория
- ✅ **Длинные названия** - обработка строк до 1000 символов

#### Результаты:
```
+ UpdateUniversityName preserves all fields except name: OK, passed 100 tests.
+ UpdateBranchName preserves all fields except name: OK, passed 100 tests.
+ UpdateFacultyName preserves all fields except name: OK, passed 100 tests.
+ UpdateGroupName preserves all fields except number: OK, passed 100 tests.
+ Name updates handle Unicode strings correctly: OK, passed 100 tests.
+ Name updates handle empty and whitespace strings: OK, passed 100 tests.
+ Updating name to same value is idempotent: OK, passed 100 tests.
+ Repository errors are properly propagated: OK, passed 100 tests.
+ Name updates handle very long names: OK, passed 100 tests.
```

### 4. Benchmark Tests
**Файл**: `internal/usecase/structure_service_name_update_benchmark_test.go`

#### Тестируемые аспекты производительности:
- ✅ **Базовая производительность** - время выполнения операций
- ✅ **Влияние длины названия** - короткие, средние, длинные, очень длинные
- ✅ **Unicode производительность** - кириллица и эмодзи
- ✅ **Параллельное выполнение** - concurrent access simulation
- ✅ **Потребление памяти** - количество аллокаций

#### Результаты производительности:
```
BenchmarkUpdateUniversityName-10                    55159    20583 ns/op    13595 B/op    143 allocs/op
BenchmarkUpdateBranchName-10                        55516    20179 ns/op    13368 B/op    140 allocs/op
BenchmarkUpdateFacultyName-10                       56239    20057 ns/op    13375 B/op    140 allocs/op
BenchmarkUpdateGroupName-10                         54685    20703 ns/op    13579 B/op    145 allocs/op
BenchmarkUpdateUniversityName_UnicodeRussian-10     55094    20453 ns/op    13883 B/op    143 allocs/op
BenchmarkUpdateUniversityName_UnicodeEmoji-10       54996    20623 ns/op    13639 B/op    143 allocs/op
BenchmarkUpdateUniversityName_Parallel-10           37321    29530 ns/op    15836 B/op    177 allocs/op
```

### 5. Integration Tests
**Файл**: `integration-tests/structure_name_update_integration_test.go`

#### Покрытие:
- ✅ **End-to-end тестирование** - полный цикл HTTP запрос → ответ
- ✅ **Проверка изменений** - верификация через `/universities/{id}/structure`
- ✅ **Восстановление данных** - rollback изменений после тестов
- ✅ **Обработка ошибок** - 400, 404, 500 статусы

#### Тестируемые сценарии:
1. **Обновление университета** - изменение названия с проверкой в структуре
2. **Обновление группы** - изменение номера группы
3. **Несуществующие записи** - проверка 404 ошибок
4. **Валидация данных** - проверка 400 ошибок

## Покрытие тестами

### Функциональное покрытие:
- ✅ **HTTP Layer** - 100% endpoints покрыты
- ✅ **Use Case Layer** - 100% новых методов покрыты
- ✅ **Error Handling** - все типы ошибок протестированы
- ✅ **Edge Cases** - граничные случаи покрыты

### Типы тестирования:
- ✅ **Unit Tests** - изолированное тестирование компонентов
- ✅ **Property-Based Tests** - тестирование инвариантов
- ✅ **Benchmark Tests** - тестирование производительности
- ✅ **Integration Tests** - end-to-end тестирование

### Качество тестов:
- ✅ **Мокирование** - правильное использование mock объектов
- ✅ **Изоляция** - тесты не зависят друг от друга
- ✅ **Детерминированность** - стабильные результаты
- ✅ **Читаемость** - понятные названия и структура

## Метрики качества

### Производительность:
- **Среднее время выполнения**: ~20μs на операцию
- **Потребление памяти**: ~13-15KB на операцию
- **Количество аллокаций**: ~140-145 на операцию
- **Параллельная производительность**: стабильная при concurrent access

### Надежность:
- **Property-based тесты**: 100 случайных тестов на каждое свойство
- **Unicode поддержка**: полная поддержка кириллицы и эмодзи
- **Обработка ошибок**: все типы ошибок корректно обрабатываются
- **Edge cases**: пустые строки, длинные названия, отсутствующие связи

## Команды для запуска тестов

### Все тесты:
```bash
cd structure-service
go test -v ./internal/usecase/
go test -v ./internal/infrastructure/http/
```

### Property-based тесты:
```bash
go test -v ./internal/usecase/ -run "TestProperty"
```

### Benchmark тесты:
```bash
go test -bench=BenchmarkUpdate -benchmem ./internal/usecase/
```

### Integration тесты:
```bash
cd integration-tests
go test -v -run "TestStructureNameUpdate"
```

## Заключение

Создан comprehensive набор тестов, который обеспечивает:

1. **Высокое качество кода** - все новые методы покрыты тестами
2. **Надежность** - property-based тесты проверяют инварианты
3. **Производительность** - benchmark тесты контролируют скорость
4. **Интеграцию** - end-to-end тесты проверяют полный workflow

Все тесты проходят успешно и готовы для использования в CI/CD pipeline. Функциональность полностью протестирована и готова к продакшену.