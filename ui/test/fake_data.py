from app import application as app
import app.protobufs.decisiontrees_pb2 as pb
from protobuf_to_dict import protobuf_to_dict
import random
from argparse import ArgumentParser


def fake_config(num_trees):
    return pb.ForestConfig(
        splittingConstraints=pb.SplittingConstraints(
            maximumLevels=random.randint(4, 8),
        ),
        lossFunctionConfig=pb.LossFunctionConfig(
            lossFunction=pb.LOGIT,
        ),
        numWeakLearners=num_trees,
    )


def fake_tree(height):
    def recur(level):
        if level <= 0 or random.random() < 0.2:
            return pb.TreeNode(
                leafValue=random.random()
            )

        branch = pb.TreeNode(
            splitValue=random.random(),
            feature=random.randint(0, 100),
            left=recur(level - 1),
            right=recur(level - 1)
        )
        return branch

    return recur(height)


def fake_forest(height, num_trees):
    return pb.Forest(
        trees=[fake_tree(height) for _ in range(num_trees)]
    )


def fake_results(num_trees):
    def generator(start, end, variance):
        uniform = lambda: 2 * random.random() - 1
        return lambda i: \
            start + (end - start) * i / num_trees + uniform() * variance
    return pb.TrainingResults(epochResults=[
        pb.EpochResult(
            roc=generator(0.2, 0.9, 0.02)(i),
            calibration=generator(1.0, 1.0, 0.02)(i),
            normalizedEntropy=generator(0.9, 0.75, 0.01)(i)
        ) for i in range(num_trees)
    ])


def insert_random_forest(height, num_trees):
    serialize = lambda pb: protobuf_to_dict(pb)
    with app.app_context():
        app.mongo.db.decisiontrees.insert(
            serialize(pb.TrainingRow(
                forestConfig=fake_config(num_trees),
                forest=fake_forest(height, num_trees),
                trainingResults=fake_results(num_trees),
            ))
        )

if __name__ == "__main__":
    parser = ArgumentParser()
    parser.add_argument("--num_trees", type=int, default=5)
    parser.add_argument("--height", type=int, default=5)
    parser.add_argument("--num_rows", type=int, default=50)

    with app.app_context():
        app.mongo.db.decisiontrees.remove()

    opt = parser.parse_args()
    for _ in range(opt.num_rows):
        insert_random_forest(opt.height, opt.num_trees)
