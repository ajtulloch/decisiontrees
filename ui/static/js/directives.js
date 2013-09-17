'use strict';

/* Directives */

angular.module('decisiontreeDirectives', []).directive('treeVisualization', function() {
  // constants
  var m = [10, 10, 10, 10];
  var w = 900 - m[1] - m[3];
  var h = 600 - m[0] - m[2];

  return {
    restrict: 'E',
    scope: {
      tree: '=',
    },
    link: function (scope, element, attrs) {
      // set up initial svg object
      var vis = d3.select(element[0])
        .append("svg")
          .attr("width", w + m[1] + m[3])
          .attr("height", h + m[0] + m[2])
          .attr("class", "tree-visualization");

      scope.$watch('tree', function (newTree, oldTree) {
        // clear the elements inside of the directive
        vis.selectAll('*').remove();

        // if 'tree' is undefined, exit
        if (!newTree) {
          return;
        }

        var tree = d3.layout.tree()
          .sort(null)
          .size([w - 100, h])
          .children(function(d) {
            var result = [];
            if (d.left) {
              result.push(d.left)
            }
            if (d.right) {
              result.push(d.right)
            }
            return result.length === 0 ? null : result
          });

        // Deep copy the tree to avoid issues with JSONify'ing recusive
        // objects (d3.layout.tree does some weird modifications)
        var nodes = tree.nodes(JSON.parse(JSON.stringify(newTree)));
        var links = tree.links(nodes);

        // Offset each node
        nodes.forEach(function(d) {
          d.y += 5
        })

        var diagonal = d3.svg.diagonal()
          .projection(function(d) {
            return [d.x, d.y];
          });

        vis.selectAll("path.link")
          .data(links)
          .enter()
          .append("svg:path")
          .attr("class", "link")
          .attr("d", diagonal);

        var nodeGroup = vis.selectAll("g.node")
          .data(nodes)
          .enter()
          .append("svg:g")
          .attr("class", "node")
          .attr("transform", function(d) {
            return "translate(" + d.x + "," + d.y + ")";
          });

        nodeGroup.append("svg:circle")
          .style("fill", function(d) { 
            return "lightsteelblue";
          })
          .attr("r", 4.5);

        nodeGroup.append("svg:text")
          .attr("text-anchor", function(d) {
            return "start";
          })
          .attr("dx", function(d) {
           return 9;
          })
          .attr("dy", function(d) {
            return 3;
          })
          .text(function(d) {
            if (d.splitValue) {
              return "F:" + d.feature + " < " + d.splitValue.toPrecision(2);
            } 
            return "Prediction: " + d.leafValue.toPrecision(2)
          });

      });
    }
  }
});
