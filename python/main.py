import os
import json
import logging
import pathlib
import uuid
from fastapi import FastAPI, Form, HTTPException, Path
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "images"
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)


@app.get("/")
def root():
    return {"message": "Hello, world!"}

def load_items_from_json():
    items_path = pathlib.Path(__file__).parent.resolve() / "items.json"
    try:
        with open(items_path, "r") as file:
            items = json.load(file)
    except FileNotFoundError:
        items = []
    return items


def save_items_to_json(items):
    items_path = pathlib.Path(__file__).parent.resolve() / "items.json"
    with open(items_path, "w") as file:
        json.dump(items, file, indent=2)
        
@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...)):
    logger.info(f"Received item: {name}")
    items = load_items_from_json()
#made unique ID?
    new_id = str(uuid.uuid4())

    new_item = {"id": new_id, "name":name, "category":category}
    items.append(new_item)
    save_items_to_json(items)
    return {"message": f"item received: {name}", "item_id": new_id}




@app.get("/items")
def read_items():
    return {"message": "Listing items"}

@app.get("/image/{image_name}")
async def get_image(image_name):
    # Create image path
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)

@app.get("/items/{item_id}")
def read_item(item_id: int = Path(..., title="The ID of the item you want to retrieve")):
    items = load_items_from_json()  
    item = next((item for item in items if item.get("id") == item_id), None)

    if item is None:
        raise HTTPException(status_code=404, detail="Item not found")

    return {"item_id": item_id, "details": item}
