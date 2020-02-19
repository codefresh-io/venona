let files = [];

const recursive = async (path, ignore, cb) => {
	return cb(null, files);
};

recursive.__setFiles = (names) => {
	files = files.concat(names);
};

recursive.__clear = () => {
	files = [];
};

module.exports = recursive;