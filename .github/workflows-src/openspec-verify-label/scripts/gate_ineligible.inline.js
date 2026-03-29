const reason = '${{ steps.verify_label.outputs.label_verified }}' !== 'true'
  ? '${{ steps.verify_label.outputs.label_reason }}'
  : '${{ steps.select_change.outputs.selection_reason }}';

core.info(`Run is ineligible — skipping agent: ${reason}`);
core.setFailed(`Ineligible: ${reason}`);
