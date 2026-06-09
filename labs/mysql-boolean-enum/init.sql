CREATE TABLE products (
                          id INT PRIMARY KEY,
                          name VARCHAR(100) NOT NULL,
                          price INT NOT NULL
);

CREATE TABLE users (
                       id INT PRIMARY KEY AUTO_INCREMENT,
                       username VARCHAR(100) NOT NULL,
                       email VARCHAR(255) NOT NULL,
                       role VARCHAR(50) NOT NULL
);

CREATE TABLE secrets (
                         id INT PRIMARY KEY AUTO_INCREMENT,
                         value VARCHAR(255) NOT NULL
);

INSERT INTO products (id, name, price) VALUES
                                           (1, 'Laptop', 1000),
                                           (2, 'Mouse', 50),
                                           (3, 'Keyboard', 120);

INSERT INTO users (username, email, role) VALUES
                                              ('admin', 'admin@example.local', 'admin'),
                                              ('john', 'john@example.local', 'user'),
                                              ('alice', 'alice@example.local', 'user');

INSERT INTO secrets (value) VALUES
    ('flag{mysql_boolean_enum_training}');