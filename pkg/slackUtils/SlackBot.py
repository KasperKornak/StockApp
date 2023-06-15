from flask import Flask, request, jsonify
from slack_sdk import WebClient
from slack_sdk.errors import SlackApiError
from dotenv import load_dotenv
import os


app = Flask(__name__)
load_dotenv()
slack_token = os.environ['SLACK_BOT_TOKEN']
client = WebClient(token=slack_token)

@app.route('/', methods=['POST'])
def post_data():
    content = request.json
    ticker = content['ticker']
    divUSD = content['divUSD']
    divPLN = content['divPLN']

    message = f"Ticker: {ticker}\nDividend in USD: {divUSD}\nDividend in PLN: {divPLN}"
    channel_id = "general"

    try:
        response = client.chat_postMessage(
            channel=channel_id,
            text=message
        )
    except SlackApiError as e:
        assert e.response["error"] 

    return jsonify(content), 200

if __name__ == "__main__":
    app.run(host='localhost', port=9009)
