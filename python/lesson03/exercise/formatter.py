from flask import Flask, request

app = Flask(__name__)
 
@app.route("/format")
def format():
    hello_to = request.args.get('helloTo')
    return f'Hello, {hello_to}!'
 
if __name__ == "__main__":
    app.run(port=8081)
