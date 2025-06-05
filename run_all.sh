#!/bin/bash
# Script para executar todas as máquinas com redirecionamento de logs

# Limpar logs antigos
rm -f alice_log.txt bob_log.txt carol_log.txt

# Iniciar as máquinas em terminais separados
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    osascript -e 'tell app "Terminal" to do script "cd '$(pwd)' && ./run_machine.sh alice.txt"'
    osascript -e 'tell app "Terminal" to do script "cd '$(pwd)' && ./run_machine.sh bob.txt"'
    osascript -e 'tell app "Terminal" to do script "cd '$(pwd)' && ./run_machine.sh carol.txt"'
else
    # Linux
    gnome-terminal -- bash -c "cd $(pwd) && ./run_machine.sh alice.txt; bash"
    gnome-terminal -- bash -c "cd $(pwd) && ./run_machine.sh bob.txt; bash"
    gnome-terminal -- bash -c "cd $(pwd) && ./run_machine.sh carol.txt; bash"
fi

echo "Todas as máquinas iniciadas. Os logs estão sendo gravados em:"
echo "- alice_log.txt"
echo "- bob_log.txt"
echo "- carol_log.txt"