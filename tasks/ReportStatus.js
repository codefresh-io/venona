const Base = require('./BaseTask');

class ReportStatus extends Base {
	async _getStatus() {
		return {
			message: 'All good',
		};
	}

	async run() {
		this.logger.info('Running task ReportStatus');
		const status = await this._getStatus();
		const res = await this.codefreshAPI.reportStatus(this.logger, status);
		this.logger.info(res);
	}
}
module.exports = ReportStatus;
