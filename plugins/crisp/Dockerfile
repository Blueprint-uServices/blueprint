FROM python:3.9-slim

# Set working directory
WORKDIR /app

# Copy all files (including CRISP directory and Python scripts)
COPY . .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Expose the FastAPI port
EXPOSE 8000

# Run the FastAPI app with Uvicorn
CMD ["uvicorn", "crisp_api:app", "--host", "0.0.0.0", "--port", "8000"] 