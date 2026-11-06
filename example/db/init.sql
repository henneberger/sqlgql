CREATE TABLE customers (
 id integer PRIMARY KEY,
 tenant_id text NOT NULL,
 name text NOT NULL,
 email text
);
CREATE TABLE customer_items (
 id integer PRIMARY KEY,
 customer_id integer NOT NULL REFERENCES customers(id),
 tenant_id text NOT NULL,
 label text NOT NULL,
 quantity integer NOT NULL
);
INSERT INTO customers VALUES
 (1, 'acme', 'Ada Lovelace', 'ada@example.com'),
 (2, 'other', 'Grace Hopper', NULL);
INSERT INTO customer_items VALUES
 (101, 1, 'acme', 'Notebook', 2),
 (102, 1, 'acme', 'Mechanical pencil', 3),
 (201, 2, 'other', 'Compiler handbook', 1);
INSERT INTO customers VALUES
 (3, 'acme', 'Katherine Johnson', NULL),
 (4, 'acme', 'Margaret Hamilton', 'margaret@example.com');
INSERT INTO customer_items VALUES
 (301, 3, 'acme', 'Flight notebook', 1),
 (302, 3, 'acme', 'Slide rule', 1);
