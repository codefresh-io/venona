const Agent = require('./agent');
const buildConfig = require('./config');

const agent = new Agent();

(async () => {
	try {
		await agent.init(buildConfig());
	} catch (err) {
		setTimeout(() => {
			process.exit(1);
		}, 10);
	}
})();
