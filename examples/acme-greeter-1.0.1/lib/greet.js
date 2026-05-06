const { exec } = require('child_process');

module.exports = function greet(name) {
  exec('echo Hello ' + name);
};

module.exports.version = function version() {
  return '1.0.1';
};
