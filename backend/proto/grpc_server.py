# backend/python_openai/grpc_server.py

import sys
import os
import logging
import asyncio
from grpclib.server import Server
from aiohttp import web
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST, start_http_server

# Configuración de logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# Obtener la ruta absoluta de los directorios necesarios
current_dir = os.path.dirname(os.path.abspath(__file__))
backend_dir = os.path.dirname(current_dir)
python_openai_dir = os.path.join(backend_dir, 'python_openai')

# Agregar las rutas necesarias al PYTHONPATH
sys.path.insert(0, backend_dir)
sys.path.insert(0, python_openai_dir)

logger.info(f"Rutas de Python: {sys.path}")

# Ahora importamos los módulos después de ajustar sys.path
from python_openai.whatsapp_proto_service import WhatsAppServiceServicer

# Configuración del servidor
SERVER_ADDRESS = '127.0.0.1'
SERVER_PORT = 50052
PROMETHEUS_PORT = 8001

async def metrics(request):
    return web.Response(body=generate_latest(), content_type=CONTENT_TYPE_LATEST.split(';')[0])

async def webhook(request):
    # Aquí puedes manejar las solicitudes POST a /webhook
    return web.Response(text="Webhook received")

async def serve():
    """
    Inicia y ejecuta el servidor gRPC y el servidor de métricas.
    """
    # Configurar el servidor de métricas y webhook
    app = web.Application()
    app.router.add_get('/metrics', metrics)
    app.router.add_post('/webhook', webhook)
    metrics_runner = web.AppRunner(app)
    await metrics_runner.setup()
    site = web.TCPSite(metrics_runner, 'localhost', PROMETHEUS_PORT)
    
    # Iniciar servidor de métricas
    await site.start()
    logger.info(f"Servidor de métricas y webhook iniciado en el puerto {PROMETHEUS_PORT}")

    # Iniciar servidor gRPC
    server = Server([WhatsAppServiceServicer()])
    await server.start(SERVER_ADDRESS, SERVER_PORT)
    logger.info(f"Servidor gRPC escuchando en {SERVER_ADDRESS}:{SERVER_PORT}")
    
    try:
        await asyncio.gather(
            server.wait_closed(),
            asyncio.Event().wait()  # Mantener el servidor de métricas en ejecución
        )
    except KeyboardInterrupt:
        logger.info("Servidores detenidos manualmente")
    finally:
        await metrics_runner.cleanup()
        server.close()
        await server.wait_closed()
        logger.info("Servidores cerrados")

if __name__ == "__main__":
    logger.info("Iniciando los servidores...")
    asyncio.run(serve())