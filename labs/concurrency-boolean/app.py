from flask import Flask, request
import sqlite3

app = Flask(__name__)


def get_db():
    conn = sqlite3.connect(":memory:")
    conn.row_factory = sqlite3.Row

    conn.execute("""
                 CREATE TABLE products (
                                           id INTEGER PRIMARY KEY,
                                           name TEXT NOT NULL,
                                           price INTEGER NOT NULL
                 )
                 """)

    conn.execute("""
                 CREATE TABLE secrets (
                                          id INTEGER PRIMARY KEY,
                                          value TEXT NOT NULL
                 )
                 """)

    conn.execute("INSERT INTO products (id, name, price) VALUES (1, 'Laptop', 1000)")
    conn.execute("INSERT INTO products (id, name, price) VALUES (2, 'Mouse', 50)")
    conn.execute("INSERT INTO secrets (id, value) VALUES (1, 'flag{tr4ining_$3cr3T_fl4g_v3r4_l0ng_t0_ext4ct}')")

    return conn


@app.route("/")
def index():
    return {
        "service": "sqli-oracle-trainer",
        "endpoints": [
            "/boolean?id=1"
        ]
    }


@app.route("/boolean")
def boolean_sqli():
    product_id = request.args.get("id", "1")

    conn = get_db()

    query = f"SELECT id, name, price FROM products WHERE id = '{product_id}'"

    try:
        row = conn.execute(query).fetchone()
    except Exception as e:
        return f"SQL error: {e}", 500

    if row:
        return "Product found", 200

    return "Not found", 200


if __name__ == "__main__":
    app.run(host="127.0.0.1", port=5006, debug=True)