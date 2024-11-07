from __future__ import print_function

# This Python file uses the following encoding: utf-8
#import logging
import time
import sys
import re

class uObject:
	""" Port del uObject a Python, woot ! """
	URBANO_VERSION = "0.000"
	debugLevel		= 0
	debugger2Log	= False
	Test			= None
	__dataMap		= {}
	# Datamap para probar
	#__dataMap		= {"test": {"property": "Test"},"hora": {"datatype": "timestamp", "default": "__time"},"uuid": {"datatype": "uuid"}}

	def setError(self, newerr):
		#uDOC:TITLE setError
		#uDOC:DESCR Fija un error en el objeto
		#uDOC:RTVAL bool false siempre 
			
		#	$this->Error_History[] = $newerr;
		#	$this->Error = $newerr;
		#	if(count($this->Error_History) > 100) { array_shift($this->Error_History); }
		self.debuggerShow(newerr, 1, "error")
		return False

	def setWarning(self, warn):
		#uDOC:TITLE setWarning
		#uDOC:DESCR Fija un warning en el uobject que puede ser sacado despues
		#uDOC:RTVAL bool false siempre

		#$this->Warning_History[] = $warn;
		#	$this->Warning = $warn;
		#	if(count($this->Warning_History) > 100) { array_shift($this->Warning_History); }
		return self.debuggerShow(warn, 1, "warn")
		
	
	def Error(self):
		#uDOC:TITLE Error 
		#uDOC:DESCR Regresa el ultimo error registrado en el uobject
		#uDOC:RTVAL	string Texto de ultimo error que paso, si no hya error ultimo null

		return self.Error
		
	def Debugger2Log(self, pon = None):
		#uDOC:TITLE Debugger2Log
		#uDOC:DESCR Pono quita la flag que hace que los errores se vayan al error_log (no al browser)

		if type(pon).__name__ == "bool":
			self.debugger2Log = pon

		return self.debugger2Log
	
	def Debug(self, level = None):
		#uDOC:TITLE Debug
		#uDOC:DESCR Lee o fija el nivel de debug de un objeto
		#uDOC:RTVAL int Nivel de debug actual (generalmente de 1 a 5)

		# Si recibimos un nivel de debug, ponlo:
		if type(level).__name__ == "int":
			self.debugLevel = level

		"""
			Si hay un constante UOBJECT_DEBUG, entonces regresa eso como debug.
			De esta forma podemos definir un nivel de debug antes de inicializar 
			un objeto (para que jale en el constructor):
		"""
		if 'UOBJECT_DEBUG' in globals():
			return globals['UOBJECT_DEBUG']
		else:
			return self.debugLevel
	
	def debuggerShow(self, message, level = 1, messagetype = "info" ):
		#uDOC:TITLE debuggerShow
		#uDOC:DESCR Saca un mensaje de debug segun el nivel requerido
		#uDOC:RTVAL bool Siempre true

		if self.debugLevel >= level:

			# Si hay $this, logueamos tambien la clase: 
			if self:
				if self.Debugger2Log() or not 'SERVER_SOFTWARE' in globals():
					classname = "{:.3f}".format(time.time()) + " " + self.__class__.__name__ + ": "
				else:
					classname = "<span style=\"color: #666;\">" + number_format(round(microtime(true),3), 3, ".", "") + " " + self.__class__.__name__ + ":</span> "
			
			# Haz el dump del recado al log o al browser, a segun:
			if self.Debugger2Log():
				error_log('[' + strtoupper(messagetype) + '] ' + classname + message)

			
			elif 'SERVER_SOFTWARE' in globals():
				# Errores al browser:
				if messagetype == "info":
					tag = "<span style=\"color:blue;font-weight:bold;\">[INFO]</span> " + classname
				elif messagetype == "warn":
					tag = "<span style=\"color:#fc0;font-weight:bold;\">[WARN]</span> " + classname
				elif messagetype == "error":
					tag = "<span style=\"color:#c33;font-weight:bold;\">[ERROR] " + classname
				
				
				print("<div style=\"font-size: 11px;font-family: monospace;\">" + tag + message + "</div>")
			
			
			# Si no hay variable SERVER_SOFTWARE es que a la mejor estamos en la consola y 
			# hay que comportarse mas como script:			
			else:
				# Errores a la consola:
				
				tag = "\033[0;37m["
				
				if messagetype == "info":
					tag = tag + "\033[1;34mINFO" + str(self.debugLevel)
				elif messagetype == "warn":
					tag = tag + "\033[1;33mWARN" + str(self.debugLevel)
				elif messagetype == "error":
					tag = tag + "\033[1;31mERROR"
				
				tag = tag + "\033[0;37m]"
				
				tag = tag + " \033[0;37m" + classname + "\033[0m"

				#logging.basicConfig(level=logging.DEBUG)
				#logging.info(tag + message)

				print(tag + message, file=sys.stderr)
				
				#print(tag + message)
				# else { print "$tag$message\n"; }
	
	def _setInstanceData(self, data):
		#uDOC:TITLE __setInstanceData
		#uDOC:DESCR Fija los datos del uObject desde un array
		
		if type(data).__name__ != 'dict':
			return self.setError("Imposible asignar datos de instancia, no recibi array de data agent")
			
		# Primero que nada, checa si viene incluido el dataagent
		if "dataagent" in data:
			self.__dataAgent = data["dataagent"]
			del(data["dataagent"])

		if "debug" in data:
			self.Debug(data['debug'])
			del(data["debug"])

		# if "sqllink" in

		if type(self.__dataMap).__name__ != "dict":
			return True
		
		for key in list(data):
			
			if key in self.__dataMap:
				# Checa si es un dataMap 2.0:
				
				if "property" in self.__dataMap[key]:
					if hasattr(self, self.__dataMap[key]["property"]):
						if type(getattr(self, self.__dataMap[key]["property"])).__name__ == "NoneType":
							setattr(self, self.__dataMap[key]["property"], data[key])
						else:
							self.setWarning("Ignoro "+key+" porque ya tiene valor")
												
						del(data[key])
				else:
					self.setWarning("Sepc para  " + key + " no tiene property en datamap. No puedo usar")

		# Si sobran warningea
		if len(data):
			if self.Debug() > 1:
				lastr = ""
				for key in data:
					lastr = lastr + key + " "
					lastr = " (" + lastr + ")"
			else:
				lastr = ""
			self.setWarning("Ojo, _setInstanceData no pudo asignar " + str(len(data)) + " elementos " + lastr)
		
		return True

	def _buildInstanceDataMap(self, input_data = None):
		#uDOC:TITLE __buildInstanceDataMap
		#uDOC:DESCR Inicializa el datamap de una instancia de este objeto
		
		datamap = self.__dataMap
		if type(datamap).__name__ != "dict":
			return self.setError("Imposible inicializar datamap, no pude obtener spec")
		
		if type(input_data).__name__ != "dict":
			input_data = {}
		
		# Si viene un debugger aqui, copia el debugger a esta instancia y borra
		# Este es un mecanismo para poder tener introspeccion en objeto efimeros
		if "debug" in input_data:
			self.Debug(input_data["debug"])
			del(input_data["debug"])
		
		data = {}
		
		for moniker in list(self.__dataMap):
			
			# A. La primera consideracion es que venga en el input data
			if moniker in input_data and input_data[moniker]:
				# Si tenemos un verificador para el tipo de dato, checamos antes
				if type(self.__dataMap[moniker]).__name__ == "dict":
					# Si la propiedad esta marcada como UUID, verifica que sea uno valido
					if "datatype" in self.__dataMap[moniker]:
						if self.__dataMap[moniker]["datatype"] == "uuid":
							if type(input_data[moniker]).__name__ != "str":
								return self.setError("Imposible usar valor de " + moniker + " como UUID, espero un string pero no es")
							if not re.match('^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$', input_data[moniker]):
								return self.setError("Imposible crear " + self.__class__.__name__ + ", " + moniker + " no es un UUID valido")

						# Si la propiedad esta marcada como array, checa que sea uno
						elif self.__dataMap[moniker]["datatype"] == "array":
							if type(input_data[moniker]).__name__ != "dict" and type(input_data[moniker]).__name__ != "array":
								return self.setError("Imposible crear map de " + self.__class__.__name__ + ", " + moniker + " no es un array")
						
						elif self.__dataMap[moniker]["datatype"] == "float" and type(input_data[moniker]).__name__ != "float":
							# FIXME: Tal vez aqui podriamos considerar castear integers... no se
							return self.setError("Imposible usar valor de " + moniker + ", spec dice float pero no es")
						
						elif self.__dataMap[moniker]["datatype"] == "float" and type(input_data[moniker]).__name__ != "int":
							# FIXME: Tal vez aqui podriamos considerar castear integers... no se
							return self.setError("Imposible usar valor de " + moniker + ", spec dice int pero no es")

						elif self.__dataMap[moniker]["datatype"] == "string" and type(input_data[moniker]).__name__ != "str":
							# FIXME: Tal vez aqui podriamos considerar castear integers... no se
							return self.setError("Imposible usar valor de " + moniker + ", spec dice string pero no es")

				# Asigna el valor al iput de regreso
				data[moniker] = input_data[moniker]
				
			# B. Si input_data no tiene definido el dato
			else:
				if type(self.__dataMap[moniker]).__name__ == "dict":
					# Pero sabemos el default:
					if "default" in self.__dataMap[moniker]:
						# Si es integer, podria ser fecha
						if "datatype" in self.__dataMap[moniker] and self.__dataMap[moniker]["datatype"] == "timestamp" and self.__dataMap[moniker]["default"] == "__time":
							 data[moniker] = int(time.time())
						else:
							data[moniker] = self.__dataMap[moniker]["default"]
					
					# elseif(isset($dataspec["required"])) { return $this->setError("Imposible crear instancia de ".get_class($this).", no recibi $moniker "); }
					else:
						if "datatype" in self.__dataMap[moniker]:
							if self.__dataMap[moniker]["datatype"] == "array":
								data[moniker] = {}
							elif self.__dataMap[moniker]["datatype"] == "int":
								data[moniker] = 0
							elif self.__dataMap[moniker]["datatype"] == "timestamp":
								data[moniker] = 0
							elif self.__dataMap[moniker]["datatype"] == "float":
								data[moniker] = 0.0
							elif self.__dataMap[moniker]["datatype"] == "string":
								data[moniker] = ""
							else: 
								data[moniker] = ""
				else:
					# En esta version de python, este no deberia de correr nunca.
					data[moniker] = ""
		
		# Regresa el array de datos
		return data
		


if __name__ == "__main__":
	from pprint import pprint			
	objeto = uObject()
	objeto.Debug(2)
	objeto.debuggerShow("Hola")
	objeto.setError("Hay problemas")
	objeto._setInstanceData({"test": "Chido"})

	pprint(objeto._buildInstanceDataMap({"test":"Test", "uuid":"8d557775-e068-49d6-8126-537a8f57416a"}))

# pprint(objeto.__dict__)

