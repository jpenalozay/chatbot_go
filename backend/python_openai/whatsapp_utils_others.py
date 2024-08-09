import asyncio
import logging
from openai import OpenAI
from dotenv import load_dotenv
import os
import time

# Cargar variables de entorno
load_dotenv()
OPENAI_API_KEY = os.getenv("OPENAI_API_KEY")
OPENAI_API_KEY_ASSISTANT = os.getenv("OPENAI_API_KEY_ASSISTANT")
OPENAI_API_KEY_ASSISTANT_ANALIZER = os.getenv("OPENAI_API_KEY_ASSISTANT_ANALIZER")
client = OpenAI(api_key=OPENAI_API_KEY)

# Configurar el registro
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logging.info("Iniciando el componente Service para Whatsapp dentro de Python...")

# Cache para el asistente
_cached_assistant = None

async def get_assistant(api_key):
    """
    Recupera y almacena en caché el asistente configurado con la API Key específica.
    
    Args:
        api_key (str): Clave de la API del asistente.
    
    Returns:
        object: Asistente recuperado o None en caso de error.
    """
    global _cached_assistant
    if _cached_assistant is None:
        logging.info(f"Recuperando el asistente con la API Key: {api_key}")
        try:
            _cached_assistant = await client.beta.assistants.retrieve(api_key)
            logging.info(f"Asistente {_cached_assistant.id} recuperado exitosamente.")
        except Exception as e:
            logging.error(f"Error al recuperar el asistente con la API Key {api_key}: {str(e)}")
            return None
    return _cached_assistant

async def execute_assistant(thread_id, api_key):
    """
    Ejecuta el asistente en un hilo de conversación y recupera el mensaje generado.
    
    Args:
        thread_id (str): ID del hilo de conversación.
        api_key (str): Clave de la API del asistente.
    
    Returns:
        str: Mensaje generado o mensaje de error.
    """
    logging.info(f"Iniciando el asistente para el hilo con ID: {thread_id}.")
    try:
        # Recuperar el asistente con caché
        assistant = await get_assistant(api_key)
        if not assistant:
            return "Lo siento, ocurrió un error al recuperar el asistente."
        
        # Iniciar ejecución del asistente
        logging.info(f"Iniciando run para el hilo con ID: {thread_id} usando el asistente con ID: {assistant.id}.")
        run = await client.beta.threads.runs.create(
            thread_id=thread_id,
            assistant_id=assistant.id
        )
        logging.info(f"Run iniciado: {run.id}, estado: {run.status}")

        # Esperar la finalización del run
        start_time = time.time()
        timeout = 30  # Segundos
        while time.time() - start_time < timeout:
            run = await client.beta.threads.runs.retrieve(thread_id=thread_id, run_id=run.id)
            logging.info(f"Verificando estado del run: {run.id}, estado: {run.status}")
            if run.status == "completed":
                logging.info(f"Run {run.id} completado exitosamente.")
                break
            await asyncio.sleep(1)
        else:
            logging.error(f"Run {run.id} para el hilo {thread_id} no se completó en el tiempo esperado.")
            return "Lo siento, ocurrió un error durante la ejecución."

        # Obtener el último mensaje generado
        messages = await client.beta.threads.messages.list(thread_id=thread_id)
        if messages.data:
            new_message = messages.data[-1].content  # Recupera el último mensaje
            logging.info(f"Mensaje generado: {new_message}")
            return new_message
        else:
            logging.warning("No se recibieron mensajes en la respuesta.")
            return "Lo siento, no se recibió una respuesta."
    except Exception as e:
        logging.error(f"Error al ejecutar el asistente para el hilo {thread_id}: {str(e)}")
        return "Lo siento, ocurrió un error al procesar tu solicitud."

async def generate_response(name, phone, thread_id, message_body, api_key):
    """
    Genera una respuesta basada en un mensaje de usuario, utilizando el asistente de OpenAI.
    
    Args:
        name (str): Nombre del usuario.
        phone (str): Número de teléfono del usuario.
        thread_id (str): ID del hilo de conversación.
        message_body (str): Mensaje del usuario.
        api_key (str): Clave de la API del asistente.
    
    Returns:
        str: Respuesta generada o mensaje de error.
    """
    logging.info(f"Generando respuesta para {phone} con hilo {thread_id}. Mensaje: {message_body}")
    try:
        # Agregar el mensaje del usuario al hilo
        logging.info(f"Agregando mensaje al hilo con ID: {thread_id}, rol: 'user'.")
        await client.beta.threads.messages.create(
            thread_id=thread_id,
            role='user',
            content=message_body
        )
        logging.info(f"Mensaje de usuario agregado al hilo {thread_id}.")
        
        # Ejecutar el asistente en el hilo para generar la respuesta
        return await execute_assistant(thread_id, api_key)
    except Exception as e:
        logging.error(f"Error al generar respuesta para el hilo {thread_id}: {str(e)}")
        return "Lo siento, ocurrió un error al procesar tu solicitud."

async def create_thread():
    """
    Crea un nuevo hilo utilizando la API de OpenAI.
    
    Returns:
        str: ID del hilo creado o None en caso de error.
    """
    logging.info("Creando un nuevo hilo.")
    try:
        thread = await client.beta.threads.create()
        logging.info(f"Hilo {thread.id} creado exitosamente.")
        return thread.id
    except Exception as e:
        logging.error(f"Error al crear el hilo: {str(e)}")
        return None

async def retrieve_thread(thread_id):
    """
    Recupera un hilo existente por su ID utilizando la API de OpenAI.
    
    Args:
        thread_id (str): ID del hilo de conversación.
    
    Returns:
        object: Hilo recuperado o None en caso de error.
    """
    logging.info(f"Recuperando el hilo con ID: {thread_id}.")
    try:
        thread = await client.beta.threads.retrieve(thread_id)
        logging.info(f"Hilo {thread.id} recuperado exitosamente.")
        return thread
    except Exception as e:
        logging.error(f"Error al recuperar el hilo {thread_id}: {str(e)}")
        return None

async def create_thread_analyzer():
    """
    Crea un nuevo hilo analizador utilizando la API de OpenAI.
    
    Returns:
        str: ID del hilo analizador creado o None en caso de error.
    """
    logging.info("Creando un nuevo hilo analizador.")
    try:
        thread = await client.beta.threads.create()
        logging.info(f"Hilo analizador {thread.id} creado exitosamente.")
        return thread.id
    except Exception as e:
        logging.error(f"Error al crear el hilo analizador: {str(e)}")
        return None

async def retrieve_thread_analyzer(thread_id):
    """
    Recupera un hilo analizador existente por su ID utilizando la API de OpenAI.
    
    Args:
        thread_id (str): ID del hilo de conversación.
    
    Returns:
        object: Hilo analizador recuperado o None en caso de error.
    """
    logging.info(f"Recuperando el hilo analizador con ID: {thread_id}.")
    try:
        thread = await client.beta.threads.retrieve(thread_id)
        logging.info(f"Hilo analizador {thread.id} recuperado exitosamente.")
        return thread
    except Exception as e:
        logging.error(f"Error al recuperar el hilo analizador {thread_id}: {str(e)}")
        return None

async def delete_thread(thread_id):
    """
    Elimina un hilo existente por su ID utilizando la API de OpenAI.
    
    Args:
        thread_id (str): ID del hilo a eliminar.
    """
    logging.info(f"Eliminando el hilo con ID: {thread_id}.")
    try:
        await client.beta.threads.delete(thread_id)
        logging.info(f"Hilo {thread_id} eliminado exitosamente.")
    except Exception as e:
        logging.error(f"Error al eliminar el hilo {thread_id}: {str(e)}")

async def delete_all_threads():
    """
    Elimina todos los hilos existentes utilizando la API de OpenAI.
    
    Returns:
        bool: True si se eliminan todos los hilos, False si ocurre un error.
    """
    logging.info("Eliminando todos los hilos.")
    try:
        threads = await list_threads()
        if threads:
            for thread in threads:
                await delete_thread(thread.id)
                logging.info(f"Hilo {thread.id} eliminado.")
        else:
            logging.info("No se encontraron hilos para eliminar.")
    except Exception as e:
        logging.error(f"Error al eliminar hilos: {str(e)}")
        return False
    return True

async def list_threads():
    """
    Obtiene todos los hilos existentes utilizando la API de OpenAI.
    
    Returns:
        list: Lista de hilos existentes o lista vacía en caso de error.
    """
    logging.info("Listando todos los hilos.")
    try:
        response = await client.beta.threads.list()
        if response and response.data:
            threads = response.data
            logging.info(f"Se listaron {len(threads)} hilos.")
            return threads
        else:
            logging.warning("No se encontraron hilos.")
            return []
    except Exception as e:
        logging.error(f"Error al listar hilos: {str(e)}")
        return []



