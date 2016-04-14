package main.com.iotracks.core_networking.local_client.private_client;

import com.iotracks.elements.IOMessage;
import io.netty.channel.Channel;
import main.com.iotracks.core_networking.local_client.LocalClient;
import main.com.iotracks.core_networking.main.CoreNetworking;
import main.com.iotracks.core_networking.utils.Constants;
import main.com.iotracks.core_networking.utils.MessageRepository;

import java.io.ByteArrayOutputStream;

/**
 * communications in "private" mode
 *
 * Created by saeid on 4/13/16.
 */
public class PrivateLocalClient implements LocalClient {
    private final long ACK_TIMEOUT = 2000;

    private Channel comSatChannel;
    private ByteArrayOutputStream buffer;
    private Object ackLock = new Object();

    /**
     * sends buffered messages to ComSat server
     *
     */
    private Runnable sendMessages = () -> {
        IOMessage message;
        while (true) {
            try {
                if (!comSatChannel.isActive())
                    break;

                while ((message = MessageRepository.peekMessage()) == null) {
                    synchronized (MessageRepository.messageLock) {
                        MessageRepository.messageLock.wait();
                    }
                }

                comSatChannel.writeAndFlush(message.getBytes()).sync();
                synchronized (ackLock) {
                    try {
                        comSatChannel.writeAndFlush(Constants.TXEND).sync();
                        ackLock.wait(ACK_TIMEOUT);
                        MessageRepository.removeHead();
                    } catch (Exception e) {
                        Thread.sleep(500);
                    }
                }
            } catch (Exception e) {
            }
        }
    };

    public PrivateLocalClient(Channel comSatChannel) {
        buffer = new ByteArrayOutputStream();
        this.comSatChannel = comSatChannel;
        new Thread(sendMessages).start();
    }

    /**
     * handle received data from ComSat server
     *
     * @param message
     */
    @Override
    public void handleMessage(byte[] message) {
        if (message.length < 6) {
            String messageString = new String(message);
            if (messageString.equals("TXEND")) {
                try {
                    IOMessage ioMessage = new IOMessage(buffer.toByteArray());
                    CoreNetworking.ioFabricClient.sendMessageToWebSocket(ioMessage);
                    buffer.reset();
                    comSatChannel.writeAndFlush(Constants.ACK).sync();
                } catch (Exception e) {
                }
                return;
            } else if (messageString.equals("ACK")) {
                synchronized (ackLock) {
                    ackLock.notifyAll();
                }
                return;
            }
        }
        try {
            buffer.write(message);
        } catch (Exception e) {
        }
    }

    @Override
    public boolean connect(long timeout) {
        return true;
    }

    @Override
    public boolean isConnected() {
        return true;
    }
}
