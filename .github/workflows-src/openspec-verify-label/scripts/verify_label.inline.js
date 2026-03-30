//include: ../../lib/verify-label.js

const result = verifyLabel(context.payload.label?.name);
core.setOutput('label_verified', result.label_verified);
core.setOutput('label_reason', result.label_reason);
core.info(result.log_message);
