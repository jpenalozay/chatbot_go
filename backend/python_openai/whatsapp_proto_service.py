# backend/python_openai/whatsapp_proto_service.py

import sys
import os
import logging
from grpclib.server import Stream
from prometheus_client import Summary, Counter

# Configurar logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# El directorio actual ya contiene los archivos proto generados
current_dir = os.path.dirname(os.path.abspath(__file__))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

logger.info(f"Rutas de Python en whatsapp_proto_service: {sys.path}")

try:
    import whatsapp_pb2
    import whatsapp_grpc
except ImportError as e:
    logger.error(f"Error al importar módulos proto: {e}")
    raise

from whatsapp_utils import (
    create_thread,
    create_thread_analyzer,
    generate_response,
    generate_response_analyzer
)

# Métricas de Prometheus
REQUEST_TIME = Summary('grpc_request_processing_seconds', 'Time spent processing gRPC request', ['method'])
REQUEST_COUNT = Counter('grpc_request_count', 'Number of gRPC requests processed', ['method'])

class WhatsAppServiceServicer(whatsapp_grpc.WhatsAppServiceBase):
    
    @REQUEST_COUNT.labels(method='CreateThread').count_exceptions()
    @REQUEST_TIME.labels(method='CreateThread').time()
    async def CreateThread(self, stream: Stream):
        request = await stream.recv_message()
        logger.info("Creando hilo...")
        try:
            thread_id = await create_thread()
            response = whatsapp_pb2.CreateThreadResponse(thread_id=thread_id)
            logger.info(f"Hilo creado con ID: {thread_id}")
        except Exception as e:
            logger.error(f"Error al crear el hilo: {str(e)}")
            response = whatsapp_pb2.CreateThreadResponse(thread_id="")
        await stream.send_message(response)

    @REQUEST_COUNT.labels(method='CreateThreadAnalizer').count_exceptions()
    @REQUEST_TIME.labels(method='CreateThreadAnalizer').time()
    async def CreateThreadAnalizer(self, stream: Stream):
        request = await stream.recv_message()
        logger.info("Creando hilo analizador...")
        try:
            thread_id_analyzer = await create_thread_analyzer()
            response = whatsapp_pb2.CreateThreadAnalizerResponse(thread_id_analizer=thread_id_analyzer)
            logger.info(f"Hilo analizador creado con ID: {thread_id_analyzer}")
        except Exception as e:
            logger.error(f"Error al crear el hilo analizador: {str(e)}")
            response = whatsapp_pb2.CreateThreadAnalizerResponse(thread_id_analizer="")
        await stream.send_message(response)

    @REQUEST_COUNT.labels(method='GenerateResponse').count_exceptions()
    @REQUEST_TIME.labels(method='GenerateResponse').time()
    async def GenerateResponse(self, stream: Stream):
        request = await stream.recv_message()
        logger.info(f"Generando respuesta para el hilo {request.thread_id}...")
        try:
            response_text = await generate_response(
                request.phone, request.thread_id, request.message_body
            )
            response = whatsapp_pb2.GenerateResponseResponse(response=response_text)
            logger.info(f"Respuesta generada para el hilo {request.thread_id}")
        except Exception as e:
            logger.error(f"Error al generar la respuesta: {str(e)}")
            response = whatsapp_pb2.GenerateResponseResponse(response="")
        await stream.send_message(response)

    @REQUEST_COUNT.labels(method='GenerateResponseAnalizer').count_exceptions()
    @REQUEST_TIME.labels(method='GenerateResponseAnalizer').time()
    async def GenerateResponseAnalizer(self, stream: Stream):
        request = await stream.recv_message()
        logger.info(f"Generando respuesta analizada para el hilo {request.thread_id_analizer}...")
        try:
            response_text = await generate_response_analyzer(
                request.thread_id_analizer, request.message_body
            )
            response = whatsapp_pb2.GenerateResponseAnalizerResponse(response=response_text)
            logger.info(f"Respuesta analizada generada para el hilo {request.thread_id_analizer}")
        except Exception as e:
            logger.error(f"Error al generar la respuesta analizada: {str(e)}")
            response = whatsapp_pb2.GenerateResponseAnalizerResponse(response="")
        await stream.send_message(response)