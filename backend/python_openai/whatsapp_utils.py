# backend/python_openai/whatsapp_utils.py

import asyncio
import logging
from openai import AsyncOpenAI
from dotenv import load_dotenv
import os
import time
from prometheus_client import Summary, Counter
from functools import wraps

# Configuración de logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# Cargar variables de entorno
load_dotenv()
OPENAI_API_KEY = os.getenv("OPENAI_API_KEY")
OPENAI_API_KEY_ASSISTANT = os.getenv("OPENAI_API_KEY_ASSISTANT")
OPENAI_API_KEY_ASSISTANT_ANALIZER = os.getenv("OPENAI_API_KEY_ASSISTANT_ANALIZER")

# Inicializar cliente de OpenAI
client = AsyncOpenAI(api_key=OPENAI_API_KEY)

# Métricas de Prometheus
FUNCTION_TIME = Summary('whatsapp_function_processing_seconds', 'Time spent processing WhatsApp function', ['function'])
FUNCTION_COUNT = Counter('whatsapp_function_count', 'Number of WhatsApp function calls', ['function'])

# Caché para asistentes
assistant_cache = {}

def async_timed_prometheus(func):
    @wraps(func)
    @FUNCTION_COUNT.labels(function=func.__name__).count_exceptions()
    @FUNCTION_TIME.labels(function=func.__name__).time()
    async def wrapped(*args, **kwargs):
        logger.info(f"Iniciando {func.__name__}")
        try:
            result = await func(*args, **kwargs)
            logger.info(f"Finalizando {func.__name__}")
            return result
        except Exception as e:
            logger.error(f"Error en {func.__name__}: {str(e)}")
            raise
    return wrapped

@async_timed_prometheus
async def get_assistant(api_key):
    if api_key not in assistant_cache:
        logger.info(f"Recuperando el asistente con la API Key: {api_key}")
        try:
            assistant = await client.beta.assistants.retrieve(api_key)
            assistant_cache[api_key] = assistant
            logger.info(f"Asistente {assistant.id} recuperado y cacheado exitosamente")
        except Exception as e:
            logger.error(f"Error al recuperar el asistente con la API Key {api_key}: {str(e)}")
            return None
    return assistant_cache.get(api_key)

@async_timed_prometheus
async def retrieve_thread(thread_id):
    logger.info(f"Recuperando el hilo con ID: {thread_id}")
    try:
        thread = await client.beta.threads.retrieve(thread_id)
        logger.info(f"Hilo {thread.id} recuperado exitosamente")
        return thread
    except Exception as e:
        logger.error(f"Error al recuperar el hilo {thread_id}: {str(e)}")
        return None

@async_timed_prometheus
async def retrieve_run_status(thread_id, run_id):
    logger.info(f"Recuperando el estado del run {run_id} para el hilo {thread_id}")
    try:
        run = await client.beta.threads.runs.retrieve(thread_id=thread_id, run_id=run_id)
        logger.info(f"Estado del run {run.id}: {run.status}")
        return run
    except Exception as e:
        logger.error(f"Error al recuperar el estado del run {run_id}: {str(e)}")
        return None

@async_timed_prometheus
async def wait_for_run_completion(thread_id, run_id, timeout=30):
    logger.info(f"Esperando la finalización del run {run_id} para el hilo {thread_id}")
    try:
        async with asyncio.timeout(timeout):
            while True:
                run = await retrieve_run_status(thread_id, run_id)
                if run and run.status == "completed":
                    logger.info(f"Run {run_id} completado exitosamente")
                    return True
                await asyncio.sleep(1)
    except asyncio.TimeoutError:
        logger.error(f"Run {run_id} para el hilo {thread_id} no se completó en el tiempo esperado")
        return False

@async_timed_prometheus
async def execute_assistant(thread_id, api_key):
    logger.info(f"Iniciando el asistente para el hilo con ID: {thread_id}")
    try:
        assistant = await get_assistant(api_key)
        if not assistant:
            logger.error("No se pudo recuperar el asistente")
            return "Lo siento, ocurrió un error al recuperar el asistente."
        
        run = await client.beta.threads.runs.create(thread_id=thread_id, assistant_id=assistant.id)
        logger.info(f"Run iniciado: {run.id}, estado: {run.status}")

        completed = await wait_for_run_completion(thread_id, run.id)
        if not completed:
            logger.error("La ejecución del asistente no se completó en el tiempo esperado")
            return "Lo siento, ocurrió un error durante la ejecución."

        messages = await client.beta.threads.messages.list(thread_id=thread_id)
        if messages.data:
            new_message = messages.data[0].content[0].text.value
            logger.info(f"Mensaje generado: {new_message}")
            return new_message
        else:
            logger.warning("No se recibieron mensajes en la respuesta")
            return "Lo siento, no se recibió una respuesta."
    except Exception as e:
        logger.error(f"Error al ejecutar el asistente para el hilo {thread_id}: {str(e)}")
        return "Lo siento, ocurrió un error al procesar tu solicitud."

@async_timed_prometheus
async def add_message_to_thread(thread_id, role, content):
    logger.info(f"Agregando mensaje al hilo con ID: {thread_id}, rol: {role}")
    try:
        await client.beta.threads.messages.create(thread_id=thread_id, role=role, content=content)
        logger.info(f"Mensaje agregado exitosamente al hilo {thread_id}")
    except Exception as e:
        logger.error(f"Error al agregar mensaje al hilo {thread_id}: {str(e)}")
        raise

@async_timed_prometheus
async def process_response(thread_id, role, content, api_key):
    logger.info(f"Procesando respuesta para el hilo {thread_id}")
    await add_message_to_thread(thread_id, role, content)
    return await execute_assistant(thread_id, api_key)

@async_timed_prometheus
async def generate_response(phone, thread_id, message_body):
    logger.info(f"Generando respuesta para {phone} con hilo {thread_id}. Mensaje: {message_body}")
    try:
        return await process_response(thread_id, 'user', message_body, OPENAI_API_KEY_ASSISTANT)
    except Exception as e:
        logger.error(f"Error al generar respuesta para el hilo {thread_id}: {str(e)}")
        return "Lo siento, ocurrió un error al procesar tu solicitud."

@async_timed_prometheus
async def generate_response_analyzer(thread_id_analyzer, message_body):
    logger.info(f"Generando respuesta analizada para el hilo {thread_id_analyzer}. Mensaje: {message_body}")
    try:
        return await process_response(thread_id_analyzer, 'user', message_body, OPENAI_API_KEY_ASSISTANT_ANALIZER)
    except Exception as e:
        logger.error(f"Error al generar respuesta analizada para el hilo {thread_id_analyzer}: {str(e)}")
        return "Lo siento, ocurrió un error al procesar tu solicitud."

@async_timed_prometheus
async def create_thread():
    logger.info("Creando un nuevo hilo")
    try:
        thread = await client.beta.threads.create()
        logger.info(f"Hilo {thread.id} creado exitosamente")
        return thread.id
    except Exception as e:
        logger.error(f"Error al crear el hilo: {str(e)}")
        return None

@async_timed_prometheus
async def create_thread_analyzer():
    logger.info("Creando un nuevo hilo analizador")
    try:
        thread = await client.beta.threads.create()
        logger.info(f"Hilo analizador {thread.id} creado exitosamente")
        return thread.id
    except Exception as e:
        logger.error(f"Error al crear el hilo analizador: {str(e)}")
        return None