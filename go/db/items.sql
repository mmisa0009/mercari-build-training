
CREATE TABLE categories (
    id INTEGER PRIMARY KEY,
    name TEXT
);

INSERT INTO categories (id, name) VALUES (1, 'fashion');

CREATE TABLE items (
    id INTEGER PRIMARY KEY,
    name TEXT,
    category_id INTEGER,
    image_name TEXT,
    FOREIGN KEY (category_id) REFERENCES categories (id)
);

INSERT INTO items (name, category_id, image_name)
    VALUES ('jacket', 1, 'jacket.jpg');