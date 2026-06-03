#!/usr/bin/env python3
"""
Generate the acceptance-test PyTorch model fixture for
elasticstack_elasticsearch_ml_trained_model_deployment.

This script:
1. Downloads a tiny HuggingFace BERT fill-mask model
2. Exports it to TorchScript (a single ~17MB .pt file)
3. Writes vocabulary.json (tokenizer vocabulary)
4. Writes model_config.json (ES PUT trained model config)

Requirements: pip install torch transformers

The output files are checked into git and consumed by
EnsureFixturePyTorchModel in internal/acctest/ml_pytorch_model.go.
"""

import json
import os

import torch
import transformers

DIR = os.path.dirname(os.path.abspath(__file__))
HUB_MODEL_ID = "google/bert_uncased_L-2_H-128_A-2"
TASK_TYPE = "fill_mask"


def main():
    print(f"Loading {HUB_MODEL_ID} ({TASK_TYPE})...")
    config = transformers.AutoConfig.from_pretrained(HUB_MODEL_ID)
    config.torchscript = True
    tokenizer = transformers.AutoTokenizer.from_pretrained(HUB_MODEL_ID, use_fast=False)
    model = transformers.AutoModelForMaskedLM.from_pretrained(
        HUB_MODEL_ID, config=config
    )
    model.eval()

    # Wrapper: HuggingFace returns namedtuples like MaskedLMOutput(logits=...),
    # but the traced model must return a single tensor for ES native inference.
    class Wrap(torch.nn.Module):
        def __init__(self, m):
            super().__init__()
            self._m = m.eval()

        def forward(self, *args):
            return self._m(*args).logits

    wrapped = Wrap(model)
    wrapped.eval()

    # Dummy inputs matching what ES will send (BERT style: 4 tensors).
    text = "Who was Jim Henson? Jim Henson was a puppeteer"
    inp = tokenizer([text], padding="max_length", return_tensors="pt")
    inputs = (
        inp["input_ids"],
        inp["attention_mask"],
        torch.zeros_like(inp["input_ids"]),  # token_type_ids
        torch.arange(inp["input_ids"].size(1), dtype=torch.long),  # position_ids
    )

    print("Tracing TorchScript model...")
    traced = torch.jit.trace(wrapped, example_inputs=inputs, strict=False)
    traced = torch.jit.freeze(traced)

    pt_path = os.path.join(DIR, "traced_pytorch_model.pt")
    torch.jit.save(traced, pt_path)
    pt_size = os.path.getsize(pt_path)
    print(f"  → traced_pytorch_model.pt ({pt_size / 1024 / 1024:.1f} MB)")

    # Vocabulary (sorted by token ID).
    print("Extracting vocabulary...")
    vocab_items = tokenizer.get_vocab().items()
    vocabulary = [k for k, _ in sorted(vocab_items, key=lambda kv: kv[1])]
    with open(os.path.join(DIR, "vocabulary.json"), "w") as f:
        json.dump({"vocabulary": vocabulary}, f, separators=(",", ":"))
    print(f"  → vocabulary.json ({len(vocabulary)} tokens)")

    # Model config for PUT /_ml/trained_models.
    es_config = {
        "description": f"Tiny BERT fixture for acceptance tests ({HUB_MODEL_ID}, {TASK_TYPE})",
        "input": {"field_names": ["text_field"]},
        "inference_config": {
            "fill_mask": {
                "num_top_classes": 5,
                "mask_token": "[MASK]",
                "tokenization": {
                    "bert": {
                        "do_lower_case": True,
                        "with_special_tokens": True,
                        "max_sequence_length": 512,
                        "truncate": "first",
                        "span": -1,
                    }
                },
            }
        },
    }
    with open(os.path.join(DIR, "model_config.json"), "w") as f:
        json.dump(es_config, f, indent=2)
    print("  → model_config.json")
    print("\nDone. Regenerate by running: python " + os.path.basename(__file__))


if __name__ == "__main__":
    main()
