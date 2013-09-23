from bson.objectid import ObjectId
from flask import current_app
from flask.ext import restful


def construct_response(f):
    f['_id'] = str(f['_id'])
    return f


class DecisionTreeTask(restful.Resource):
    def get(self, task_id):
        task = current_app.mongo.db.decisiontrees.find_one(
            {"_id": ObjectId(task_id)}
        )

        if task is None:
            restful.abort(404, message="Task {} doesn't exist".format(task_id))
        return construct_response(task), 201

    def options(self):
        pass


class DecisionTreeTaskList(restful.Resource):
    def get(self):
        tasks = current_app.mongo.db.decisiontrees.find()
        return [construct_response(f) for f in tasks], 201

    def options(self):
        pass


class DecisionTreeWeakLearner(restful.Resource):
    def get(self, task_id, tree_id):
        task = current_app.mongo.db.decisiontrees.find_one(
            {"_id": ObjectId(task_id)}
        )

        if task is None:
            restful.abort(404, message="Task {} doesn't exist".format(task_id))

        return task["forest"]["trees"][tree_id], 201

    def options(self):
        pass
