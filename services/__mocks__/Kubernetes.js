// __mocks__/Kubernetes.js
const Kubernetes = require.requireActual('./../Kubernetes');
module.exports = jest.fn().mockImplementation(() => Kubernetes);
