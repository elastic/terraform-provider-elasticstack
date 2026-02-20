package datafeedstate

const (
	createTimeoutErrorMessage = "The operation to create the ML datafeed state timed out after %s." +
		" You may need to allocate more free memory within ML nodes by either closing other jobs, or increasing the overall ML memory." +
		" You may retry the operation."
	updateTimeoutErrorMessage = "The operation to update the ML datafeed state timed out after %s." +
		" You may need to allocate more free memory within ML nodes by either closing other jobs, or increasing the overall ML memory." +
		" You may retry the operation."
)
