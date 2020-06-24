import multiprocessing
import sys
import os

IDS = (
    [0,100],
    [101,200],
    [201,300],
    [301,400],
    [401,500],
    [501,600],
    [601,700],
    [701,800],
    [801, 900],
    # [901,1000],
    # [1001,1100],
    # [1101,1200],
)
NUMBER_OF_INTERFACES = 2


def create_test_namespaces(numbers):
    exec_command('modprobe ip_conntrack')
    print('Creating namespaces: %s' % numbers)
    for ns_num in range(*numbers):
        os.system('ip netns add test-router-%d' % ns_num)
        for intf_num in range(NUMBER_OF_INTERFACES):
            exec_command('ip link add test-eth-%d%d type dummy'
                         % (ns_num, intf_num))
            exec_command('ip link set test-eth-%d%d netns test-router-%d'
                         % (ns_num, intf_num, ns_num))
            exec_command('ip netns exec test-router-%d ifconfig '
                         'test-eth-%d%d up' % (ns_num, ns_num, intf_num))
        print('namespace: test-router-%d created' % ns_num)


def delete_test_namespaces(numbers):
    print('Deleting namespaces: %s' % numbers)
    for ns_num in range(*numbers):
        exec_command('ip netns del test-router-%d' % ns_num)
        for intf_num in range(NUMBER_OF_INTERFACES):
            exec_command('ip link del test-eth-%d%d'
                         % (ns_num, intf_num))


def exec_command(command):
    print("Exec: %s" % command)
    os.system(command)

if __name__ == '__main__':
    if len(sys.argv) < 3:
        print('Usage: %s [create|delete] <NUMBER_OF_THREADS>' % sys.argv[0])
    p = multiprocessing.Pool(int(sys.argv[2]))
    if sys.argv[1] == 'create':
        p.map(create_test_namespaces, IDS)
    if sys.argv[1] == 'delete':
        p.map(delete_test_namespaces, IDS)
