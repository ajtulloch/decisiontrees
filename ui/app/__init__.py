from flask import Flask, make_response
from flask.ext import restful
from flask.ext.pymongo import PyMongo
import app.decisiontrees.resources as resources
from app import constants

app = Flask(__name__, static_folder=constants.STATIC_FOLDER)
app.config['MONGO_DBNAME'] = constants.DB_NAME

app.mongo = PyMongo(app)

api = restful.Api(app)


@app.route("/")
def index():
    return make_response(open('static/index.html').read())

api.add_resource(
    resources.DecisionTreeTask,
    '/api/decisiontrees/<string:task_id>'
)
api.add_resource(resources.DecisionTreeTaskList, '/api/decisiontrees/')
