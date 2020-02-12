const Agent = require('./agent');
const buildConfig = require('./config');

if (process.argv.includes('-v')) {
	process.env.verbose = true;
}

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
