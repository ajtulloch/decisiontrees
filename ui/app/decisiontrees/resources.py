from app.protobufs.decisiontrees_pb2 import ForestConfig, Forest, TrainingRow
from bson.objectid import ObjectId
from flask import current_app
from flask.ext import restful
from protobuf_to_dict import protobuf_to_dict

def construct_response(f):
  print f
  f['_id'] = str(f['_id'])
  return f

class DecisionTreeTask(restful.Resource):
  def get(self, task_id):
    task = current_app.mongo.db.decisiontrees.find_one(
      {"_id": ObjectId(task_id)}
    )
    return construct_response(task), 201

  def options(self):
    pass

class DecisionTreeTaskList(restful.Resource):
  def get(self):
    return [construct_response(f) for f in 
            current_app.mongo.db.decisiontrees.find()], 201

  def options(self):
    pass
