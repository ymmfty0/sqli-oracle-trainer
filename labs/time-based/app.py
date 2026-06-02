from flask import Flask, request
import sqlite3
import time

app = Flask(__name__)


def sqlite_sleep(seconds):
    try:
        seconds = float(seconds)
    except Exception:
        seconds = 0

    if seconds < 0:
        seconds = 0

    if seconds > 3:
        seconds = 3

    time.sleep(seconds)
    return 0


def get_db():
    conn = sqlite3.connect(":memory:")
    conn.row_factory = sqlite3.Row

    # Register custom SQL function:
    # sleep(0.7)
    conn.create_function("sleep", 1, sqlite_sleep)

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
    conn.execute("INSERT INTO secrets (id, value) VALUES (1, 'flag{time_based_training_secret}')")

    return conn


@app.route("/")
def index():
    return {
        "service": "time-based-sqli-lab",
        "endpoints": [
            "/time?id=1"
        ],
        "examples": {
            "normal": "/time?id=1",
            "true_timing": "/time?id=1' AND CASE WHEN 1=1 THEN sleep(0.7) ELSE 0 END=0-- -",
            "false_timing": "/time?id=1' AND CASE WHEN 1=2 THEN sleep(0.7) ELSE 0 END=0-- -"
        }
    }


@app.route("/time")
def time_based_sqli():
    product_id = request.args.get("id", "1")

    conn = get_db()

    query = f"SELECT id, name, price FROM products WHERE id = '{product_id}'"

    try:
        row = conn.execute(query).fetchone()
    except Exception as e:
        return f"SQL error: {e}", 500

    # Important:
    # For time-based SQLi, the body should not be the oracle.
    # We intentionally keep the response simple.
    if row:
        return "OK", 200

    return "OK", 200


if __name__ == "__main__":
    app.run(host="127.0.0.1", port=5004, debug=True)