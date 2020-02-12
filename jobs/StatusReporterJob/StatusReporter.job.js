const Base = require('../BaseJob');

class StatusReporterJob extends Base {
	async _getStatus() {
		return {
			message: 'All good',
		};
	}

	async run() {
		const status = await this._getStatus();
		const res = await this.codefreshAPI.reportStatus(this.logger, status);
		this.logger.infoVerbose(res);
	}

	async validate() {
		return;
	}
}
module.exports = StatusReporterJob;
