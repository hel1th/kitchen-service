-- +goose Up
-- +goose StatementBegin
-- Ресторан 1: Пиццерия «Тесто и Сыр»
INSERT INTO restaurants (id, name, status, webhook_url)
VALUES ('11111111-1111-1111-1111-111111111111', 'Пиццерия «Тесто и Сыр»', 'open', 'http://restaurant-simulator:8081/webhooks/orders')
ON CONFLICT (id) DO NOTHING;

-- Категории для Ресторана 1
INSERT INTO categories (id, restaurant_id, name, sort_order)
VALUES
    ('22222222-2222-2222-2222-222222222201', '11111111-1111-1111-1111-111111111111', 'Пицца', 1),
    ('22222222-2222-2222-2222-222222222202', '11111111-1111-1111-1111-111111111111', 'Напитки', 2)
ON CONFLICT (id) DO NOTHING;

-- Блюда для Ресторана 1
INSERT INTO dishes (id, restaurant_id, category_id, name, description, price, available)
VALUES
    ('33333333-3333-3333-3333-333333333301', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222201', 'Пепперони', 'Пикантная салями, моцарелла, томатный соус', 599.00, true),
    ('33333333-3333-3333-3333-333333333302', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222201', 'Маргарита', 'Моцарелла, томаты, соус песто', 499.00, true),
    ('33333333-3333-3333-3333-333333333303', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222202', 'Лимонад 0.5', 'Освежающий цитрусовый лимонад', 120.00, true)
ON CONFLICT (id) DO NOTHING;

-- Ресторан 2: Бургерная «Мясной Крафт»
INSERT INTO restaurants (id, name, status, webhook_url)
VALUES ('11111111-1111-1111-1111-111111111112', 'Бургерная «Мясной Крафт»', 'open', 'http://restaurant-simulator:8081/webhooks/orders')
ON CONFLICT (id) DO NOTHING;

-- Категории для Ресторана 2
INSERT INTO categories (id, restaurant_id, name, sort_order)
VALUES
    ('22222222-2222-2222-2222-222222222203', '11111111-1111-1111-1111-111111111112', 'Бургеры', 1)
ON CONFLICT (id) DO NOTHING;

-- Блюда для Ресторана 2
INSERT INTO dishes (id, restaurant_id, category_id, name, description, price, available)
VALUES
    ('33333333-3333-3333-3333-333333333304', '11111111-1111-1111-1111-111111111112', '22222222-2222-2222-2222-222222222203', 'Фирменный Бургер', 'Сочная котлета из 100% говядины на гриле', 349.00, true),
    ('33333333-3333-3333-3333-333333333305', '11111111-1111-1111-1111-111111111112', '22222222-2222-2222-2222-222222222203', 'Чизбургер', 'Говядина, сыр чеддер, маринованные огурчики', 119.00, true)
ON CONFLICT (id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM dishes WHERE restaurant_id IN ('11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111112');
DELETE FROM categories WHERE restaurant_id IN ('11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111112');
DELETE FROM restaurants WHERE id IN ('11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111112');
-- +goose StatementEnd
