# PyTorch fixture model

This directory contains the tiny TorchScript model fixture used by
`EnsureFixturePyTorchModel`.

Files:

- `traced_pytorch_model.pt` — the TorchScript model artifact
- `vocabulary.json` — tokenizer vocabulary for the model
- `model_config.json` — Elasticsearch `PUT /_ml/trained_models` request body
- `export_fixture.py` — script used to regenerate the fixture from HuggingFace

The model is based on:

- `google/bert_uncased_L-2_H-128_A-2`
- task type: `fill_mask`

Regenerate with:

```bash
python3 export_fixture.py
```

Notes:

- The `.pt` file is tracked with Git LFS via the repository `.gitattributes`.
- The acceptance-test helper chunks the `.pt` file in Go and uploads it with the raw Elasticsearch ML APIs.
