import javax.jms.JMSException;
import javax.jms.Message;
import javax.jms.MessageListener;
import javax.jms.TextMessage;

public class JMSListener implements MessageListener {

    public void onMessage(Message message){
        System.out.println("## entry onMessage");

        if(message instanceof TextMessage) {
            TextMessage textMessage = (TextMessage) message;
            try {
                System.out.println("__ MyMessageListener received message with payload :" + textMessage.getText());
            } catch (JMSException jmse) {
                System.out.println("JMSException in MyMessageListener class!");
                System.out.println(jmse.getLinkedException());
            }
        } else {
            System.out.println("---Message Received was not of type TextMessage \n");
        }

        System.out.println("## exit onMessage");
    }
}
