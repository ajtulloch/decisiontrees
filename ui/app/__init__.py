from flask import Flask, make_response
from flask.ext import restful
from flask.ext.pymongo import PyMongo
import app.decisiontrees.resources as resources
from app import constants

application = Flask(__name__, static_folder=constants.STATIC_FOLDER)
application.config['MONGO_DBNAME'] = constants.DB_NAME

application.mongo = PyMongo(application)

api = restful.Api(application)


@application.route("/")
def index():
    return make_response(open('static/index.html').read())

api.add_resource(
    resources.DecisionTreeTask,
    '/api/decisiontrees/<string:task_id>'
)

api.add_resource(
    resources.DecisionTreeWeakLearner,
    '/api/decisiontrees/<string:task_id>/trees/<int:tree_id>'
)

api.add_resource(resources.DecisionTreeTaskList, '/api/decisiontrees/')
