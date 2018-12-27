const Agent = require('./agent');
const buildConfig = require('./config');

const agent = new Agent(buildConfig());

agent.init();