const { exec } = require('child_process');

module.exports = function greet(name) {
  exec('echo Hello ' + name);
};
