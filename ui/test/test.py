from app import application as app
import app.protobufs.decisiontrees_pb2 as pb
from flask.ext.pymongo import PyMongo
from protobuf_to_dict import protobuf_to_dict
import fake_data
import json
import unittest


class FlaskrTestCase(unittest.TestCase):
    def setUp(self):
        app.config['TEST_DBNAME'] = 'ui_test'
        try:
            app.mongo = PyMongo(app, config_prefix='TEST')
        except:
            pass

        self.client = app.test_client()
        num_trees, height = (5, 5)
        self.row = pb.TrainingRow(
            forestConfig=fake_data.fake_config(num_trees),
            forest=fake_data.fake_forest(height, num_trees),
        )

        with app.test_request_context():
            self._id = str(app.mongo.db.decisiontrees.insert(
                protobuf_to_dict(self.row))
            )

    def test_decision_tree_list(self):
        rv = self.client.get('/api/decisiontrees/')
        result = json.loads(rv.data)
        self.assertEqual(len(result), 1)
        self.assertDecisionTreeEqual(result[0])

    def assertDecisionTreeEqual(self, response):
        self.assertEqual(response["_id"], self._id)
        self.assertEqual(
            response["forestConfig"],
            protobuf_to_dict(self.row.forestConfig)
        )
        self.assertEqual(response["forest"], protobuf_to_dict(self.row.forest))

    def test_decision_tree_detail(self):
        rv = self.client.get('/api/decisiontrees/{0}'.format(self._id))
        result = json.loads(rv.data)
        self.assertDecisionTreeEqual(result)

    def test_decision_tree_nonexistent(self):
        rv = self.client.get('/api/decisiontrees/{0}'.format(0))
        result = json.loads(rv.data)
        self.assertEqual(result['status'], 500)

    def tearDown(self):
        with app.test_request_context():
            app.mongo.db.decisiontrees.remove()

if __name__ == '__main__':
    unittest.main()
