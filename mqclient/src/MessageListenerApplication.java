import com.ibm.msg.client.jms.JmsConnectionFactory;
import com.ibm.msg.client.jms.JmsFactoryFactory;
import com.ibm.msg.client.wmq.WMQConstants;

import javax.jms.*;
import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;

public class MessageListenerApplication implements MessageListener {

    private static final String HOST = "localhost";
    private static final int PORT = 1414;
    private static final String CHANNEL = "DEV.APP.SVRCONN";
    private static final String QMGR = "QM1";
    private static final String APP_USER = "app";
    private static final String APP_PASSWORD = "_APP_PASSWORD_";
    private static final String QUEUE_NAME = "LUKE.ALIAS.QUEUE";

    public static void main(String[] args) {
        JMSContext context;
        Destination destination;
        JMSConsumer consumer;

        JmsConnectionFactory connectionFactory = createJMSConnectionFactory();

        setJMSProperties(connectionFactory);

        System.out.println("MQ Test: Connecting to " + HOST + ", PORT: " + PORT + ", CHANNEL: " + CHANNEL + ",Connecting to " + QUEUE_NAME);

        context = connectionFactory.createContext();
        destination = context.createQueue("queue:///" + QUEUE_NAME);
        consumer = context.createConsumer(destination);

        // Create a Listener object, and associates the listener to the consumer
        JMSListener ml = new JMSListener();
        consumer.setMessageListener(ml);

        // Message listener will now listen for messages in a separate thread
        System.out.println("The message listener is running");

        // Startup UI
        userInterface(context, connectionFactory, destination);
    }

    private static JmsConnectionFactory createJMSConnectionFactory() {
        JmsFactoryFactory ff;
        JmsConnectionFactory cf;
        try {
            ff = JmsFactoryFactory.getInstance(WMQConstants.WMQ_PROVIDER);
            cf = ff.createConnectionFactory();
        }
        catch (JMSException jmse) {
            System.out.println("JMS Exception when trying to instantiate connection factory");
            if (jmse.getLinkedException() != null) {
                System.out.println(jmse.getLinkedException());
            } else {jmse.printStackTrace();}
            cf = null;
        }
        return cf;
    }

    private static void setJMSProperties(JmsConnectionFactory cf) {
        try {
            cf.setStringProperty(WMQConstants.WMQ_HOST_NAME, HOST);
            cf.setIntProperty(WMQConstants.WMQ_PORT, PORT);
            cf.setStringProperty(WMQConstants.WMQ_CHANNEL, CHANNEL);
            cf.setIntProperty(WMQConstants.WMQ_CONNECTION_MODE, WMQConstants.WMQ_CM_CLIENT);
            cf.setStringProperty(WMQConstants.WMQ_QUEUE_MANAGER, QMGR);
            cf.setStringProperty(WMQConstants.WMQ_APPLICATIONNAME, "JmsPutGet (JMS)");
            cf.setBooleanProperty(WMQConstants.USER_AUTHENTICATION_MQCSP, true);
            cf.setStringProperty(WMQConstants.USERID, APP_USER);
            cf.setStringProperty(WMQConstants.PASSWORD, APP_PASSWORD);
        } catch (JMSException jmse) {
            System.out.println("JMS Exception when trying to set JMS properties!");
            if (jmse.getLinkedException() != null){ // if there is an associated linked exception, print it. Otherwise print the stack trace
                System.out.println(jmse.getLinkedException());
            } else {jmse.printStackTrace();}
        }
    }

    public static void userInterface(JMSContext context, JmsConnectionFactory connectionFactory, Destination destination) {
        BufferedReader br = new BufferedReader(new InputStreamReader(System.in));
        boolean exit = false;
        while (!exit) {
            String command;
            try {
                System.out.println("Ready : ");
                command = br.readLine();
                command = command.toLowerCase();

                switch (command) {
                    case "start": case "restart":
                        context.start();
                        System.out.println("--Message Listener Started --");
                        break;
                    case "stop":
                        context.stop();
                        System.out.println("--Message Listener Stopped--");
                        break;
                    case "send":
                        sendATextMessage(connectionFactory, destination);
                        System.out.println("--Sent Message--");
                        break;
                    case "exit":
                        context.close();
                        System.out.println("--Exiting --");
                        exit = true;
                        break;
                    default:
                        System.out.println("Help: Valid commands are start/restart, stop, send, and exit");
                }
            } catch (IOException e) {
                e.printStackTrace();
            }
        }
        System.exit(0);
    }

    public static void sendATextMessage(JmsConnectionFactory connectionFactory, Destination destination) {
        try {
            BufferedReader br = new BufferedReader(new InputStreamReader(System.in));
            System.out.println("Payload : ");
            String payload = br.readLine();

            JMSContext producerContext = connectionFactory.createContext();
            JMSProducer producer = producerContext.createProducer();
            TextMessage m = producerContext.createTextMessage(payload);
            producer.send(destination, m);
            producerContext.close();
        } catch (Exception e) {
            System.out.println("Exception sending message!");
            e.printStackTrace();
        }
    }

    @Override
    public void onMessage(javax.jms.Message message) {

    }
}
