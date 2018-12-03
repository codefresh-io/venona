const _ = require('lodash');

function getPropertyOrError(obj, path, errorMessage) {
	const result = _.get(obj, path);
	if (!result) {
		throw new Error(errorMessage);
	}
	return result;
}

module.exports = {
	getPropertyOrError,
};
