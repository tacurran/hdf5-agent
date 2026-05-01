from flask import Flask, jsonify, request
from flask_cors import CORS
import h5py
import json
import os

app = Flask(__name__)
CORS(app, resources={r"/api/*": {"origins": "*"}})

@app.route('/api/health', methods=['GET'])
def health():
    return jsonify({"status": "healthy", "backend": "Python Flask with h5py"})

@app.route('/api/files', methods=['GET'])
def list_files():
    files = []
    for f in os.listdir('.'):
        if f.endswith(('.h5', '.hdf5')):
            size = os.path.getsize(f)
            files.append({"name": f, "size": size, "path": f})
    return jsonify(files)

@app.route('/api/file/structure', methods=['GET'])
def get_structure():
    path = request.args.get('path')
    if not path:
        return jsonify({"error": "Missing path"}), 400
    
    try:
        with h5py.File(path, 'r') as f:
            def build_tree(group, name):
                items = []
                for key in group.keys():
                    item = group[key]
                    if isinstance(item, h5py.Group):
                        items.append({
                            "name": key,
                            "path": item.name,
                            "type": "group",
                            "children": build_tree(item, item.name)
                        })
                    elif isinstance(item, h5py.Dataset):
                        items.append({
                            "name": key,
                            "path": item.name,
                            "type": "dataset",
                            "shape": list(item.shape),
                            "dtype": str(item.dtype)
                        })
                return items
            
            return jsonify({
                "name": os.path.basename(path),
                "path": path,
                "type": "file",
                "children": build_tree(f, "/")
            })
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/api/dataset/read', methods=['GET'])
def read_dataset():
    file_path = request.args.get('file')
    dataset_path = request.args.get('path')
    
    if not file_path or not dataset_path:
        return jsonify({"error": "Missing parameters"}), 400
    
    try:
        with h5py.File(file_path, 'r') as f:
            dataset = f[dataset_path]
            data = dataset[()].tolist() if dataset.size < 100000 else dataset[:100].tolist()
            
            return jsonify({
                "name": dataset.name,
                "path": dataset_path,
                "data": data,
                "shape": list(dataset.shape),
                "dtype": str(dataset.dtype)
            })
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    app.run(port=8080, debug=False)