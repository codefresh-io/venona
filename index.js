const Agent = require('./agent');
const buildConfig = require('./config');

const agent = new Agent(buildConfig());

(async () => {
	try {
		await agent.init();
	} catch (err) {
		setTimeout(() => {
			process.exit(1);
		}, 10);
	}
})();
