# Import using a bare parameter UUID (default space; state id becomes default/<parameter_uuid>)
terraform import elasticstack_kibana_synthetics_parameter.my_param <parameter_uuid>

# Or using the composite <space_id>/<parameter_uuid> form
terraform import elasticstack_kibana_synthetics_parameter.my_param <space_id>/<parameter_uuid>
