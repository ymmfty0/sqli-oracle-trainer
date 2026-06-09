from flask import Flask, request
import pymysql

app = Flask(__name__)


def get_db():
    return pymysql.connect(
        host="127.0.0.1",
        port=3307,
        user="appuser",
        password="apppass",
        database="appdb",
        cursorclass=pymysql.cursors.DictCursor,
        autocommit=True,
    )


@app.route("/")
def index():
    return {
        "service": "mysql-boolean-enum-lab",
        "endpoints": [
            "/boolean?id=1"
        ],
        "examples": {
            "true": "/boolean?id=1' AND 1=1-- -",
            "false": "/boolean?id=1' AND 1=2-- -"
        }
    }


@app.route("/boolean")
def boolean_sqli():
    product_id = request.args.get("id", "1")

    query = f"SELECT id, name, price FROM products WHERE id = '{product_id}'"

    try:
        conn = get_db()
        with conn.cursor() as cursor:
            cursor.execute(query)
            row = cursor.fetchone()
    except Exception as e:
        return f"SQL error: {e}", 500
    finally:
        try:
            conn.close()
        except Exception:
            pass

    if row:
        return "Product found", 200

    return "Not found", 200


if __name__ == "__main__":
    app.run(host="127.0.0.1", port=5007, debug=True)